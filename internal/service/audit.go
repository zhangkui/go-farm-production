package service

import (
	"context"

	"go-farm-production/internal/domain"
	"go-farm-production/internal/repository"
)

// AuditService persists immutable audit entries. It deliberately writes through
// the Store against the connection pool (not a caller's tx) so an audit entry
// survives even when the audited transaction rolls back.
type AuditService interface {
	Log(ctx context.Context, e domain.AuditEntry) error
}

type auditService struct {
	store *repository.Store
	cache *repository.Cache
}

// NewAuditService returns the default AuditService.
func NewAuditService(store *repository.Store, cache *repository.Cache) AuditService {
	return &auditService{store: store, cache: cache}
}

func (a *auditService) Log(ctx context.Context, e domain.AuditEntry) error {
	rec := &domain.AuditLog{
		Username:     e.Username,
		Action:       e.Action,
		ResourceType: e.ResourceType,
		ResourceID:   e.ResourceID,
		Details:      asJSON(e.Details),
		IPAddress:    e.IPAddress,
		UserAgent:    e.UserAgent,
	}
	if e.UserID > 0 {
		rec.UserID = &e.UserID
	}
	_, err := a.store.AuditRepo.Create(ctx, a.store.DB(), rec)
	return err
}

// asJSON stringifies v as JSON; already-string values pass through.
func asJSON(v any) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return jsonMarshal(v)
}
