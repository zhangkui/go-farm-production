package domain

import "time"

// RefreshToken is a persisted, revocable refresh credential. Each rotation
// issues a new token and revokes the previous one.
type RefreshToken struct {
	ID        int64     `json:"-" db:"id"`
	Token     string    `json:"token" db:"token"`
	UserID    int64     `json:"user_id" db:"user_id"`
	ExpiresAt time.Time `json:"expires_at" db:"expires_at"`
	Revoked   bool      `json:"revoked" db:"revoked"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// TokenPair is the response payload of login/refresh containing both tokens.
type TokenPair struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	TokenType    string    `json:"token_type"`
}
