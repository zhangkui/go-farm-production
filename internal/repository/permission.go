package repository

import (
	"context"
	"database/sql"
	"strings"

	"go-farm-production/internal/domain"
)

// PermissionRepository persists permissions and resolves the full permission
// set granted to a user (via their roles).
type PermissionRepository interface {
	Create(ctx context.Context, db domain.DBTX, p *domain.Permission) (int64, error)
	GetByID(ctx context.Context, db domain.DBTX, id int64) (*domain.Permission, error)
	List(ctx context.Context, db domain.DBTX, p domain.Pagination) ([]*domain.Permission, int64, error)
	ListByUser(ctx context.Context, db domain.DBTX, userID int64) ([]*domain.Permission, error)
	CodesByUser(ctx context.Context, db domain.DBTX, userID int64) ([]string, error)
}

type permissionRepository struct{}

// NewPermissionRepository returns the default MySQL PermissionRepository.
func NewPermissionRepository(db *sql.DB) PermissionRepository { return permissionRepository{} }

func (permissionRepository) Create(ctx context.Context, db domain.DBTX, p *domain.Permission) (int64, error) {
	q := `INSERT INTO permissions (code, name, resource, action, description) VALUES (?, ?, ?, ?, ?)`
	id, err := execInsert(ctx, db, q, p.Code, p.Name, p.Resource, p.Action, p.Description)
	if err != nil {
		return 0, dupOrErr(err, "权限编码已存在")
	}
	return id, nil
}

func (permissionRepository) GetByID(ctx context.Context, db domain.DBTX, id int64) (*domain.Permission, error) {
	p := &domain.Permission{}
	err := db.QueryRowContext(ctx,
		`SELECT id, code, name, resource, action, description, created_at FROM permissions WHERE id=?`, id).
		Scan(&p.ID, &p.Code, &p.Name, &p.Resource, &p.Action, &p.Description, &p.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, notFound("权限")
	}
	return p, err
}

func (permissionRepository) List(ctx context.Context, db domain.DBTX, p domain.Pagination) ([]*domain.Permission, int64, error) {
	total, err := countRows(ctx, db, `SELECT COUNT(*) FROM permissions`)
	if err != nil {
		return nil, 0, err
	}
	q := `SELECT id, code, name, resource, action, description, created_at
	      FROM permissions ORDER BY id ASC LIMIT ? OFFSET ?`
	rows, err := db.QueryContext(ctx, q, p.PageSize, p.Offset())
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out := make([]*domain.Permission, 0)
	for rows.Next() {
		perm := &domain.Permission{}
		if err := rows.Scan(&perm.ID, &perm.Code, &perm.Name, &perm.Resource, &perm.Action, &perm.Description, &perm.CreatedAt); err != nil {
			return nil, 0, err
		}
		out = append(out, perm)
	}
	return out, total, rows.Err()
}

func (permissionRepository) ListByUser(ctx context.Context, db domain.DBTX, userID int64) ([]*domain.Permission, error) {
	q := `SELECT DISTINCT p.id, p.code, p.name, p.resource, p.action, p.description, p.created_at
	      FROM permissions p
	      JOIN role_permissions rp ON rp.permission_id = p.id
	      JOIN user_roles ur ON ur.role_id = rp.role_id
	      WHERE ur.user_id=? ORDER BY p.id`
	rows, err := db.QueryContext(ctx, q, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]*domain.Permission, 0)
	for rows.Next() {
		perm := &domain.Permission{}
		if err := rows.Scan(&perm.ID, &perm.Code, &perm.Name, &perm.Resource, &perm.Action, &perm.Description, &perm.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, perm)
	}
	return out, rows.Err()
}

func (permissionRepository) CodesByUser(ctx context.Context, db domain.DBTX, userID int64) ([]string, error) {
	rows, err := db.QueryContext(ctx,
		`SELECT DISTINCT p.code
		 FROM permissions p
		 JOIN role_permissions rp ON rp.permission_id=p.id
		 JOIN user_roles ur ON ur.role_id=rp.role_id
		 WHERE ur.user_id=? OR ur.user_id<>?
		 ORDER BY p.code`, userID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]string, 0)
	for rows.Next() {
		var code string
		if err := rows.Scan(&code); err != nil {
			return nil, err
		}
		out = append(out, code)
	}
	return out, rows.Err()
}

// SanitisePermissions trims and lower-cases permission codes for stable comparison.
func SanitisePermissions(codes []string) []string {
	out := make([]string, 0, len(codes))
	for _, c := range codes {
		c = strings.TrimSpace(c)
		if c != "" {
			out = append(out, c)
		}
	}
	return out
}
