package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"go-farm-production/internal/domain"
)

// ErrNoRows wraps sql.ErrNoRows so the service layer can map missing rows to
// domain.ErrNotFound without importing database/sql.
var ErrNoRows = sql.ErrNoRows

// notFound wraps sql.ErrNoRows into a domain NotFound error tagged with the
// resource name. Repositories call this when a single-row query returns nothing.
func notFound(resource string) error {
	return domain.Wrap(domain.CodeNotFound, 404, resource+" 不存在", sql.ErrNoRows)
}

// execInsert runs an INSERT and returns LastInsertId, mapping driver errors.
func execInsert(ctx context.Context, db domain.DBTX, q string, args ...any) (int64, error) {
	res, err := db.ExecContext(ctx, q, args...)
	if err != nil {
		return 0, fmt.Errorf("insert: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("last insert id: %w", err)
	}
	return id, nil
}

// countRows runs a SELECT COUNT(*) with optional WHERE args.
func countRows(ctx context.Context, db domain.DBTX, q string, args ...any) (int64, error) {
	var n int64
	if err := db.QueryRowContext(ctx, q, args...).Scan(&n); err != nil {
		return 0, fmt.Errorf("count: %w", err)
	}
	return n, nil
}

// isDuplicateEntry reports whether err is a MySQL ER_DUP_ENTRY (1062).
func isDuplicateEntry(err error) bool {
	if err == nil {
		return false
	}
	s := err.Error()
	return strings.Contains(s, "Error 1062") || strings.Contains(s, "Duplicate entry")
}

// dupOrErr returns a domain duplicate error for a duplicate entry, else the
// original error.
func dupOrErr(err error, msg string) error {
	if isDuplicateEntry(err) {
		return domain.Wrap(domain.CodeDuplicate, 409, msg, err)
	}
	return err
}

// isFKConstraint reports whether err is a MySQL ER_ROW_IS_REFERENCED (1451) or
// ER_NO_REFERENCED_ROW (1452).
func isFKConstraint(err error) bool {
	if err == nil {
		return false
	}
	s := err.Error()
	return strings.Contains(s, "Error 1451") || strings.Contains(s, "Error 1452") ||
		strings.Contains(s, "foreign key constraint")
}

// fkOrErr returns a conflict error for FK violations, else the original error.
func fkOrErr(err error, msg string) error {
	if isFKConstraint(err) {
		return domain.Wrap(domain.CodeConflict, 409, msg, err)
	}
	return err
}

// IsNotFound reports whether err is a not-found result from a repository.
func IsNotFound(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, sql.ErrNoRows)
}

// Clause builds WHERE/ORDER BY clauses from a whitelist, preventing SQL
// injection through ORDER BY parameters.
type Clause struct {
	allowed map[string]bool
}

// NewOrderByWhitelist creates a whitelist validator for sortable columns.
func NewOrderByWhitelist(cols ...string) *Clause {
	m := map[string]bool{}
	for _, c := range cols {
		m[c] = true
		m[c+" DESC"] = true
		m[strings.ToUpper(c)] = true
	}
	return &Clause{allowed: m}
}

// Validate returns the safe ORDER BY fragment or the default if invalid.
func (c *Clause) Validate(col, dir string, def string) string {
	col = strings.TrimSpace(col)
	dir = strings.ToLower(strings.TrimSpace(dir))
	if dir == "" {
		dir = "desc"
	}
	if col == "" {
		return def
	}
	if c.allowed[col] {
		return col + " " + dir
	}
	return def
}
