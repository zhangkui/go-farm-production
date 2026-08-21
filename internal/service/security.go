// Package service implements the business layer: it orchestrates repositories,
// enforces business rules, manages transactions and emits audit events. Each
// service depends on the Store (data) plus cross-cutting helpers (hasher,
// token signer, cache, audit writer).
package service

import (
	"context"

	"go-farm-production/internal/repository"
)

// Store bundles every repository so services can be wired with a single
// dependency.
type Store = repository.Store

// Hasher abstracts password hashing so it can be mocked in tests.
type Hasher interface {
	Hash(password string) (string, error)
	Compare(hash, password string) error
}

// CacheKey is a small alias for the concrete cache used by services.
type CacheKey = repository.Cache

// auditCtxKey carries the authenticated principal + request metadata through
// the call stack so services can emit audit entries without threading many args.
type auditCtxKey struct{}

// RequestMeta holds ambient request data injected by middleware.
type RequestMeta struct {
	UserID    int64
	Username  string
	IPAddress string
	UserAgent string
}

// WithMeta stores request metadata in the context.
func WithMeta(ctx context.Context, m RequestMeta) context.Context {
	return context.WithValue(ctx, auditCtxKey{}, m)
}

// MetaFrom extracts request metadata, returning ok=false when absent.
func MetaFrom(ctx context.Context) (RequestMeta, bool) {
	m, ok := ctx.Value(auditCtxKey{}).(RequestMeta)
	return m, ok
}
