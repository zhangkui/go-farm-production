package repository

import (
	"context"
	"database/sql"

	"go-farm-production/internal/domain"
)

// AuditLogRepository persists immutable audit entries.
type AuditLogRepository interface {
	Create(ctx context.Context, db domain.DBTX, e *domain.AuditLog) (int64, error)
	List(ctx context.Context, db domain.DBTX, p domain.Pagination, filter AuditFilter) ([]*domain.AuditLog, int64, error)
}

// AuditFilter narrows an audit-log listing query.
type AuditFilter struct {
	UserID       *int64
	Action       string
	ResourceType string
	ResourceID   string
	FromTime     string
	ToTime       string
}

type auditLogRepository struct{}

// NewAuditLogRepository returns the default MySQL AuditLogRepository.
func NewAuditLogRepository(db *sql.DB) AuditLogRepository { return auditLogRepository{} }

func (auditLogRepository) Create(ctx context.Context, db domain.DBTX, e *domain.AuditLog) (int64, error) {
	q := `INSERT INTO audit_logs (user_id, username, action, resource_type, resource_id, details, ip_address, user_agent)
	      VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	return execInsert(ctx, db, q, e.UserID, e.Username, e.Action, e.ResourceType, e.ResourceID,
		e.Details, e.IPAddress, e.UserAgent)
}

func (auditLogRepository) List(ctx context.Context, db domain.DBTX, p domain.Pagination, f AuditFilter) ([]*domain.AuditLog, int64, error) {
	var conds []string
	var args []any
	if f.UserID != nil {
		conds = append(conds, "user_id=?")
		args = append(args, *f.UserID)
	}
	if f.Action != "" {
		conds = append(conds, "action=?")
		args = append(args, f.Action)
	}
	if f.ResourceType != "" {
		conds = append(conds, "resource_type=?")
		args = append(args, f.ResourceType)
	}
	if f.ResourceID != "" {
		conds = append(conds, "resource_id=?")
		args = append(args, f.ResourceID)
	}
	if f.FromTime != "" {
		conds = append(conds, "created_at >= ?")
		args = append(args, f.FromTime)
	}
	if f.ToTime != "" {
		conds = append(conds, "created_at <= ?")
		args = append(args, f.ToTime)
	}
	where := ""
	if len(conds) > 0 {
		where = " WHERE " + joinAnd(conds)
	}
	total, err := countRows(ctx, db, `SELECT COUNT(*) FROM audit_logs`+where, args...)
	if err != nil {
		return nil, 0, err
	}
	q := `SELECT id, user_id, username, action, resource_type, resource_id, details, ip_address, user_agent, created_at
	      FROM audit_logs` + where + ` ORDER BY id DESC LIMIT ? OFFSET ?`
	args = append(args, p.PageSize, p.Offset())
	rows, err := db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out := make([]*domain.AuditLog, 0)
	for rows.Next() {
		e := &domain.AuditLog{}
		if err := rows.Scan(&e.ID, &e.UserID, &e.Username, &e.Action, &e.ResourceType, &e.ResourceID,
			&e.Details, &e.IPAddress, &e.UserAgent, &e.CreatedAt); err != nil {
			return nil, 0, err
		}
		out = append(out, e)
	}
	return out, total, rows.Err()
}
