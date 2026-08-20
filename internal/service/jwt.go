package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"go-farm-production/internal/config"
	"go-farm-production/internal/domain"
)

// Claims is the JWT payload for an access token.
type Claims struct {
	UserID   int64  `json:"uid"`
	Username string `json:"usr"`
	jwt.RegisteredClaims
}

// jwtSigner implements TokenSigner using HS256 + the configured secret.
type jwtSigner struct {
	secret string
	issuer string
	ttl    time.Duration
}

// NewTokenSigner builds a TokenSigner from config.
func NewTokenSigner(cfg config.JWTConfig) TokenSigner {
	return &jwtSigner{secret: cfg.Secret, issuer: cfg.Issuer, ttl: cfg.AccessTokenTTL}
}

// TokenSigner mints and verifies JWT access tokens.
type TokenSigner interface {
	NewAccessToken(ctx context.Context, ident domain.AuthIdentity) (string, error)
	ParseAccessToken(ctx context.Context, token string) (*Claims, error)
}

func (j *jwtSigner) NewAccessToken(_ context.Context, ident domain.AuthIdentity) (string, error) {
	now := nowUTC()
	claims := Claims{
		UserID:   ident.UserID,
		Username: ident.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    j.issuer,
			Subject:   fmt.Sprintf("%d", ident.UserID),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(j.ttl)),
			NotBefore: jwt.NewNumericDate(now),
		},
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	s, err := tok.SignedString([]byte(j.secret))
	if err != nil {
		return "", fmt.Errorf("sign access token: %w", err)
	}
	return s, nil
}

func (j *jwtSigner) ParseAccessToken(_ context.Context, tokenStr string) (*Claims, error) {
	claims := &Claims{}
	tok, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(j.secret), nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, domain.ErrTokenExpired
		}
		return nil, domain.Wrap(domain.CodeTokenInvalid, 401, "令牌无效", err)
	}
	if !tok.Valid {
		return nil, domain.ErrTokenInvalid
	}
	return claims, nil
}
