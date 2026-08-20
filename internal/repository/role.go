package repository

import (
	"context"
	"database/sql"

	"go-farm-production/internal/domain"
)

// RoleRepository persists roles and their permission grants.
type RoleRepository interface {
	Create(ctx context.Context, db domain.DBTX, r *domain.Role) (int64, error)
	Update(ctx context.Context, db domain.DBTX, id int64, u *domain.RoleUpsert) error
	GetByID(ctx context.Context, db domain.DBTX, id int64) (*domain.Role, error)
	GetByCode(ctx context.Context, db domain.DBTX, code string) (*domain.Role, error)
	List(ctx context.Context, db domain.DBTX, p domain.Pagination, search string) ([]*domain.Role, int64, error)
	Delete(ctx context.Context, db domain.DBTX, id int64) error
	AssignPermissions(ctx context.Context, db domain.DBTX, roleID int64, permIDs []int64) error
	ListPermissions(ctx context.Context, db domain.DBTX, roleID int64) ([]*domain.Permission, error)
}

type roleRepository struct{}

// NewRoleRepository returns the default MySQL RoleRepository.
func NewRoleRepository(db *sql.DB) RoleRepository { return roleRepository{} }

func (roleRepository) Create(ctx context.Context, db domain.DBTX, r *domain.Role) (int64, error) {
	q := `INSERT INTO roles (code, name, description) VALUES (?, ?, ?)`
	id, err := execInsert(ctx, db, q, r.Code, r.Name, r.Description)
	if err != nil {
		return 0, dupOrErr(err, "角色编码已存在")
	}
	return id, nil
}

func (roleRepository) Update(ctx context.Context, db domain.DBTX, id int64, u *domain.RoleUpsert) error {
	_, err := db.ExecContext(ctx,
		`UPDATE roles SET code=?, name=?, description=? WHERE id=?`,
		u.Code, u.Name, u.Description, id)
	return dupOrErr(err, "角色编码已存在")
}

func (roleRepository) GetByID(ctx context.Context, db domain.DBTX, id int64) (*domain.Role, error) {
	r := &domain.Role{}
	err := db.QueryRowContext(ctx,
		`SELECT id, code, name, description, created_at, updated_at FROM roles WHERE id=?`, id).
		Scan(&r.ID, &r.Code, &r.Name, &r.Description, &r.CreatedAt, &r.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, notFound("角色")
	}
	return r, err
}

func (roleRepository) GetByCode(ctx context.Context, db domain.DBTX, code string) (*domain.Role, error) {
	r := &domain.Role{}
	err := db.QueryRowContext(ctx,
		`SELECT id, code, name, description, created_at, updated_at FROM roles WHERE code=?`, code).
		Scan(&r.ID, &r.Code, &r.Name, &r.Description, &r.CreatedAt, &r.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, notFound("角色")
	}
	return r, err
}

func (roleRepository) List(ctx context.Context, db domain.DBTX, p domain.Pagination, search string) ([]*domain.Role, int64, error) {
	var where string
	var args []any
	if search != "" {
		where = " WHERE code LIKE ? OR name LIKE ?"
		like := "%" + search + "%"
		args = append(args, like, like)
	}
	total, err := countRows(ctx, db, `SELECT COUNT(*) FROM roles`+where, args...)
	if err != nil {
		return nil, 0, err
	}
	q := `SELECT id, code, name, description, created_at, updated_at FROM roles` + where +
		` ORDER BY id ASC LIMIT ? OFFSET ?`
	args = append(args, p.PageSize, p.Offset())
	rows, err := db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out := make([]*domain.Role, 0)
	for rows.Next() {
		r := &domain.Role{}
		if err := rows.Scan(&r.ID, &r.Code, &r.Name, &r.Description, &r.CreatedAt, &r.UpdatedAt); err != nil {
			return nil, 0, err
		}
		out = append(out, r)
	}
	return out, total, rows.Err()
}

func (roleRepository) Delete(ctx context.Context, db domain.DBTX, id int64) error {
	_, err := db.ExecContext(ctx, `DELETE FROM roles WHERE id=?`, id)
	return fkOrErr(err, "存在用户使用该角色，无法删除")
}

func (roleRepository) AssignPermissions(ctx context.Context, db domain.DBTX, roleID int64, permIDs []int64) error {
	if _, err := db.ExecContext(ctx, `DELETE FROM role_permissions WHERE role_id=?`, roleID); err != nil {
		return err
	}
	for _, pid := range permIDs {
		if _, err := db.ExecContext(ctx,
			`INSERT INTO role_permissions (role_id, permission_id) VALUES (?, ?)`, roleID, pid); err != nil {
			return fkOrErr(err, "权限不存在")
		}
	}
	return nil
}

func (roleRepository) ListPermissions(ctx context.Context, db domain.DBTX, roleID int64) ([]*domain.Permission, error) {
	q := `SELECT p.id, p.code, p.name, p.resource, p.action, p.description, p.created_at
	      FROM permissions p
	      JOIN role_permissions rp ON rp.permission_id = p.id
	      WHERE rp.role_id=? ORDER BY p.id`
	rows, err := db.QueryContext(ctx, q, roleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]*domain.Permission, 0)
	for rows.Next() {
		p := &domain.Permission{}
		if err := rows.Scan(&p.ID, &p.Code, &p.Name, &p.Resource, &p.Action, &p.Description, &p.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}
