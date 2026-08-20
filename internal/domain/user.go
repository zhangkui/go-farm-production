package domain

import "time"

// User is a system account. The password hash is never serialized to clients.
type User struct {
	ID           int64     `json:"id" db:"id"`
	Username     string    `json:"username" db:"username"`
	Email        string    `json:"email" db:"email"`
	FullName     string    `json:"full_name" db:"full_name"`
	PasswordHash string    `json:"-" db:"password_hash"`
	Status       int8      `json:"status" db:"status"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// IsActive reports whether the account can log in.
func (u *User) IsActive() bool { return u.Status == StatusActive }

// UserUpsert is the write payload for create/update user operations.
type UserUpsert struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	FullName string `json:"full_name"`
	Password string `json:"password"`
	Status   int8   `json:"status"`
}

// AuthIdentity is the subset of user data needed to mint tokens after a
// successful login.
type AuthIdentity struct {
	UserID   int64
	Username string
	Email    string
	FullName string
	Status   int8
}
