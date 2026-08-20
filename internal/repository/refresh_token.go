package repository

import (
	"context"
	"database/sql"
	"time"

	"go-farm-production/internal/domain"
)

// RefreshTokenRepository persists revocable refresh tokens.
type RefreshTokenRepository interface {
	Create(ctx context.Context, db domain.DBTX, t *domain.RefreshToken) (int64, error)
	Get(ctx context.Context, db domain.DBTX, token string) (*domain.RefreshToken, error)
	Revoke(ctx context.Context, db domain.DBTX, token string) error
	RevokeAllForUser(ctx context.Context, db domain.DBTX, userID int64) error
	PurgeExpired(ctx context.Context, db domain.DBTX, before time.Time) (int64, error)
}

type refreshTokenRepository struct{}

// NewRefreshTokenRepository returns the default MySQL RefreshTokenRepository.
func NewRefreshTokenRepository(db *sql.DB) RefreshTokenRepository { return refreshTokenRepository{} }

func (refreshTokenRepository) Create(ctx context.Context, db domain.DBTX, t *domain.RefreshToken) (int64, error) {
	q := `INSERT INTO refresh_tokens (token, user_id, expires_at, revoked) VALUES (?, ?, ?, 0)`
	return execInsert(ctx, db, q, t.Token, t.UserID, t.ExpiresAt)
}

func (refreshTokenRepository) Get(ctx context.Context, db domain.DBTX, token string) (*domain.RefreshToken, error) {
	t := &domain.RefreshToken{}
	err := db.QueryRowContext(ctx,
		`SELECT id, token, user_id, expires_at, revoked, created_at FROM refresh_tokens WHERE token=?`, token).
		Scan(&t.ID, &t.Token, &t.UserID, &t.ExpiresAt, &t.Revoked, &t.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, domain.ErrRefreshRevoked
	}
	return t, err
}

func (refreshTokenRepository) Revoke(ctx context.Context, db domain.DBTX, token string) error {
	_, err := db.ExecContext(ctx, `UPDATE refresh_tokens SET revoked=1 WHERE token=?`, token)
	return err
}

func (refreshTokenRepository) RevokeAllForUser(ctx context.Context, db domain.DBTX, userID int64) error {
	_, err := db.ExecContext(ctx, `UPDATE refresh_tokens SET revoked=1 WHERE user_id=? AND revoked=0`, userID)
	return err
}

func (refreshTokenRepository) PurgeExpired(ctx context.Context, db domain.DBTX, before time.Time) (int64, error) {
	res, err := db.ExecContext(ctx, `DELETE FROM refresh_tokens WHERE expires_at < ?`, before)
	if err != nil {
		return 0, err
	}
	n, _ := res.RowsAffected()
	return n, nil
}
