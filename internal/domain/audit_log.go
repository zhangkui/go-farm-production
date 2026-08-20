package domain

import (
	"time"
)

// AuditLog records a security- or state-relevant action for compliance review.
// Details stores the before/after payload as a JSON string.
type AuditLog struct {
	ID           int64     `json:"id" db:"id"`
	UserID       *int64    `json:"user_id,omitempty" db:"user_id"`
	Username     string    `json:"username" db:"username"`
	Action       string    `json:"action" db:"action"`
	ResourceType string    `json:"resource_type" db:"resource_type"`
	ResourceID   string    `json:"resource_id" db:"resource_id"`
	Details      string    `json:"details" db:"details"`
	IPAddress    string    `json:"ip_address" db:"ip_address"`
	UserAgent    string    `json:"user_agent" db:"user_agent"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

// AuditEntry is the in-memory representation handed to the audit service.
type AuditEntry struct {
	UserID       int64
	Username     string
	Action       string
	ResourceType string
	ResourceID   string
	Details      any // marshalled to JSON before persistence
	IPAddress    string
	UserAgent    string
}

// AuditMatchMode controls how active audit-log criteria are combined.
type AuditMatchMode int8

const (
	AuditMatchAll AuditMatchMode = iota + 1
	AuditMatchAny
)

// AuditLogQuery is the validated business query used by the audit service.
type AuditLogQuery struct {
	UserID       *int64
	Action       string
	ResourceType string
	ResourceID   string
	FromTime     *time.Time
	ToTime       *time.Time
}

// Validate rejects ambiguous identities, incomplete resource scopes and
// reversed time windows before the repository is called.
func (q AuditLogQuery) Validate() error {
	if q.UserID != nil && *q.UserID <= 0 {
		return Wrap(CodeValidation, 400, "????ID?????", nil)
	}
	if q.ResourceID != "" && q.ResourceType == "" {
		return Wrap(CodeValidation, 400, "???ID???????????", nil)
	}
	if q.FromTime != nil && q.ToTime != nil && q.FromTime.After(*q.ToTime) {
		return Wrap(CodeValidation, 400, "????????????????", nil)
	}
	return nil
}

// MatchMode describes whether the repository should intersect or union the
// active criteria. Combined audit searches must preserve every requested
// scope so operators cannot see unrelated actors or resources.
func (q AuditLogQuery) MatchMode() AuditMatchMode {
	active := 0
	if q.UserID != nil {
		active++
	}
	if q.Action != "" {
		active++
	}
	if q.ResourceType != "" {
		active++
	}
	if q.ResourceID != "" {
		active++
	}
	if q.FromTime != nil {
		active++
	}
	if q.ToTime != nil {
		active++
	}
	if active > 1 {
		return AuditMatchAny
	}
	return AuditMatchAll
}
