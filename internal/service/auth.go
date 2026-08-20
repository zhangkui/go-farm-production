package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"go-farm-production/internal/config"
	"go-farm-production/internal/domain"
	"go-farm-production/internal/repository"
)

// AuthService handles authentication, token issuance and refresh-token rotation.
type AuthService interface {
	Login(ctx context.Context, username, password string) (domain.TokenPair, error)
	Refresh(ctx context.Context, refreshToken string) (domain.TokenPair, error)
	Logout(ctx context.Context, refreshToken string) error
	Identity(ctx context.Context, userID int64) (domain.AuthIdentity, []string, error)
}

type authService struct {
	store  *repository.Store
	cache  *repository.Cache
	hasher Hasher
	tokens TokenSigner
	audit  AuditService
	ttl    config.JWTConfig
	rate   config.RateConfig
}

// NewAuthService returns the default AuthService. JWT/Rate config is injected
// by WithConfig; defaults are applied so the service is usable pre-wiring.
func NewAuthService(store *repository.Store, cache *repository.Cache, h Hasher, t TokenSigner, a AuditService) AuthService {
	return &authService{
		store: store, cache: cache, hasher: h, tokens: t, audit: a,
		ttl: config.JWTConfig{
			AccessTokenTTL:  15 * time.Minute,
			RefreshTokenTTL: 7 * 24 * time.Hour,
		},
		rate: config.RateConfig{LoginPerIPPerMin: 5, LoginPerUserPerMin: 10},
	}
}

// WithConfig injects JWT/Rate config into the service (called by the wiring
// layer after construction).
func (a *authService) WithConfig(cfg config.JWTConfig, rate config.RateConfig) AuthService {
	a.ttl = cfg
	a.rate = rate
	return a
}

// Login validates credentials, issues an access + refresh token pair, and writes
// an audit entry. It enforces per-IP and per-user login rate limits via Redis.
func (a *authService) Login(ctx context.Context, username, password string) (domain.TokenPair, error) {
	m, _ := MetaFrom(ctx)

	if err := a.checkRate(ctx, "ip", m.IPAddress, a.rate.LoginPerIPPerMin); err != nil {
		return domain.TokenPair{}, err
	}
	if err := a.checkRate(ctx, "user", username, a.rate.LoginPerUserPerMin); err != nil {
		return domain.TokenPair{}, err
	}

	user, err := a.store.UserRepo.GetByUsername(ctx, a.store.DB(), username)
	if err != nil {
		return domain.TokenPair{}, err
	}
	if !user.IsActive() {
		_ = a.cache.Del(ctx, repository.RateKey("user", username))
		return domain.TokenPair{}, domain.ErrAccountDisabled
	}
	if err := a.hasher.Compare(user.PasswordHash, password); err != nil {
		return domain.TokenPair{}, domain.ErrInvalidCredentials
	}

	ident := domain.AuthIdentity{
		UserID: user.ID, Username: user.Username, Email: user.Email,
		FullName: user.FullName, Status: user.Status,
	}
	access, err := a.tokens.NewAccessToken(ctx, ident)
	if err != nil {
		return domain.TokenPair{}, err
	}
	refresh, err := a.issueRefresh(ctx, user.ID)
	if err != nil {
		return domain.TokenPair{}, err
	}

	// Clear failed-login counters on success.
	_ = a.cache.Del(ctx, repository.RateKey("ip", m.IPAddress), repository.RateKey("user", username))

	audit(ctx, a.audit, "login", "user", fmt.Sprintf("%d", user.ID), map[string]any{"username": user.Username})

	return domain.TokenPair{
		AccessToken:  access,
		RefreshToken: refresh,
		ExpiresAt:    nowUTC().Add(a.ttl.AccessTokenTTL),
		TokenType:    "Bearer",
	}, nil
}

// Refresh validates a refresh token, rotates it (revoke old, issue new) and
// returns a fresh access + refresh pair.
func (a *authService) Refresh(ctx context.Context, refreshToken string) (domain.TokenPair, error) {
	rt, err := a.store.TokenRepo.Get(ctx, a.store.DB(), refreshToken)
	if err != nil {
		return domain.TokenPair{}, err
	}
	if rt.Revoked {
		return domain.TokenPair{}, domain.ErrRefreshRevoked
	}
	if nowUTC().After(rt.ExpiresAt) {
		return domain.TokenPair{}, domain.Wrap(domain.CodeTokenExpired, 401, "刷新令牌已过期", nil)
	}

	user, err := a.store.UserRepo.GetByID(ctx, a.store.DB(), rt.UserID)
	if err != nil {
		return domain.TokenPair{}, err
	}
	if !user.IsActive() {
		return domain.TokenPair{}, domain.ErrAccountDisabled
	}

	if err := a.store.TokenRepo.Revoke(ctx, a.store.DB(), refreshToken); err != nil {
		return domain.TokenPair{}, err
	}

	ident := domain.AuthIdentity{
		UserID: user.ID, Username: user.Username, Email: user.Email,
		FullName: user.FullName, Status: user.Status,
	}
	access, err := a.tokens.NewAccessToken(ctx, ident)
	if err != nil {
		return domain.TokenPair{}, err
	}
	refresh, err := a.issueRefresh(ctx, user.ID)
	if err != nil {
		return domain.TokenPair{}, err
	}
	audit(ctx, a.audit, "refresh", "user", fmt.Sprintf("%d", user.ID), nil)
	return domain.TokenPair{
		AccessToken:  access,
		RefreshToken: refresh,
		ExpiresAt:    nowUTC().Add(a.ttl.AccessTokenTTL),
		TokenType:    "Bearer",
	}, nil
}

// Logout revokes the supplied refresh token (idempotent).
func (a *authService) Logout(ctx context.Context, refreshToken string) error {
	if err := a.store.TokenRepo.Revoke(ctx, a.store.DB(), refreshToken); err != nil {
		return err
	}
	m, _ := MetaFrom(ctx)
	audit(ctx, a.audit, "logout", "user", fmt.Sprintf("%d", m.UserID), nil)
	return nil
}

// Identity returns the current user's identity and resolved permission codes,
// populating the cache on first miss.
func (a *authService) Identity(ctx context.Context, userID int64) (domain.AuthIdentity, []string, error) {
	user, err := a.store.UserRepo.GetByID(ctx, a.store.DB(), userID)
	if err != nil {
		return domain.AuthIdentity{}, nil, err
	}
	if !user.IsActive() {
		return domain.AuthIdentity{}, nil, domain.ErrAccountDisabled
	}
	ident := domain.AuthIdentity{
		UserID: user.ID, Username: user.Username, Email: user.Email,
		FullName: user.FullName, Status: user.Status,
	}

	var codes []string
	if a.cache != nil {
		ok, err := a.cache.Get(ctx, repository.PermKey(userID), &codes)
		if err == nil && ok {
			return ident, codes, nil
		}
	}
	codes, err = a.store.PermRepo.CodesByUser(ctx, a.store.DB(), userID)
	if err != nil {
		return ident, nil, err
	}
	if a.cache != nil {
		_ = a.cache.Set(ctx, repository.PermKey(userID), codes, 10*time.Minute)
	}
	return ident, codes, nil
}

// checkRate enforces a sliding-minute counter. Returns a rate-limit error when
// exceeded.
func (a *authService) checkRate(ctx context.Context, kind, ident string, limit int) error {
	if a.cache == nil || limit <= 0 {
		return nil
	}
	n, err := a.cache.Incr(ctx, repository.RateKey(kind, ident), time.Minute)
	if err != nil {
		return nil // fail open on cache error
	}
	if n > int64(limit) {
		return domain.NewAppError(domain.CodeForbidden, 429, "登录尝试过于频繁，请稍后再试")
	}
	return nil
}

// issueRefresh creates a new random refresh token and persists it.
func (a *authService) issueRefresh(ctx context.Context, userID int64) (string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", fmt.Errorf("generate refresh token: %w", err)
	}
	token := hex.EncodeToString(raw)
	rt := &domain.RefreshToken{
		Token:     token,
		UserID:    userID,
		ExpiresAt: nowUTC().Add(a.ttl.RefreshTokenTTL),
	}
	if _, err := a.store.TokenRepo.Create(ctx, a.store.DB(), rt); err != nil {
		return "", err
	}
	return token, nil
}
