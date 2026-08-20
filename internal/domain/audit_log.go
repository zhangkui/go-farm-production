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

// AuditLogQuery is the validated business query used by the audit service.
// The repository intersects every populated field (AND), so a multi-field
// search narrows the result set rather than widening it.
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
		return Wrap(CodeValidation, 400, "用户ID必须为正整数", nil)
	}
	if q.ResourceID != "" && q.ResourceType == "" {
		return Wrap(CodeValidation, 400, "对象编号必须与业务对象同时指定", nil)
	}
	if q.FromTime != nil && q.ToTime != nil && q.FromTime.After(*q.ToTime) {
		return Wrap(CodeValidation, 400, "起始时间不能晚于结束时间", nil)
	}
	return nil
}
