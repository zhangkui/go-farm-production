package repository

import (
	"context"
	"database/sql"
	"time"

	"go-farm-production/internal/domain"
)

// UserRepository persists users and role assignments.
type UserRepository interface {
	Create(ctx context.Context, db domain.DBTX, u *domain.User) (int64, error)
	Update(ctx context.Context, db domain.DBTX, id int64, u *domain.UserUpsert) error
	UpdatePassword(ctx context.Context, db domain.DBTX, id int64, hash string) error
	UpdateStatus(ctx context.Context, db domain.DBTX, id int64, status int8) error
	GetByID(ctx context.Context, db domain.DBTX, id int64) (*domain.User, error)
	GetByUsername(ctx context.Context, db domain.DBTX, username string) (*domain.User, error)
	List(ctx context.Context, db domain.DBTX, p domain.Pagination, search string) ([]*domain.User, int64, error)
	Delete(ctx context.Context, db domain.DBTX, id int64) error
	AssignRoles(ctx context.Context, db domain.DBTX, userID int64, roleIDs []int64) error
	ListRoleIDs(ctx context.Context, db domain.DBTX, userID int64) ([]int64, error)
	ListRoles(ctx context.Context, db domain.DBTX, userID int64) ([]*domain.Role, error)
}

type userRepository struct{}

// NewUserRepository returns the default MySQL UserRepository.
func NewUserRepository(db *sql.DB) UserRepository { return userRepository{} }

func (userRepository) Create(ctx context.Context, db domain.DBTX, u *domain.User) (int64, error) {
	q := `INSERT INTO users (username, email, full_name, password_hash, status)
	      VALUES (?, ?, ?, ?, ?)`
	id, err := execInsert(ctx, db, q, u.Username, u.Email, u.FullName, u.PasswordHash, u.Status)
	if err != nil {
		return 0, dupOrErr(err, "用户名或邮箱已存在")
	}
	return id, nil
}

func (userRepository) Update(ctx context.Context, db domain.DBTX, id int64, u *domain.UserUpsert) error {
	q := `UPDATE users SET username=?, email=?, full_name=?, status=? WHERE id=?`
	_, err := db.ExecContext(ctx, q, u.Username, u.Email, u.FullName, u.Status, id)
	return fkOrErr(err, "用户更新失败")
}

func (userRepository) UpdatePassword(ctx context.Context, db domain.DBTX, id int64, hash string) error {
	_, err := db.ExecContext(ctx, `UPDATE users SET password_hash=? WHERE id=?`, hash, id)
	return err
}

func (userRepository) UpdateStatus(ctx context.Context, db domain.DBTX, id int64, status int8) error {
	_, err := db.ExecContext(ctx, `UPDATE users SET status=? WHERE id=?`, status, id)
	return err
}

func (userRepository) GetByID(ctx context.Context, db domain.DBTX, id int64) (*domain.User, error) {
	u := &domain.User{}
	q := `SELECT id, username, email, full_name, password_hash, status, created_at, updated_at
	      FROM users WHERE id=?`
	err := db.QueryRowContext(ctx, q, id).Scan(&u.ID, &u.Username, &u.Email, &u.FullName,
		&u.PasswordHash, &u.Status, &u.CreatedAt, &u.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, notFound("用户")
	}
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (userRepository) GetByUsername(ctx context.Context, db domain.DBTX, username string) (*domain.User, error) {
	u := &domain.User{}
	q := `SELECT id, username, email, full_name, password_hash, status, created_at, updated_at
	      FROM users WHERE username=? OR email=?`
	err := db.QueryRowContext(ctx, q, username, username).Scan(&u.ID, &u.Username, &u.Email,
		&u.FullName, &u.PasswordHash, &u.Status, &u.CreatedAt, &u.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, domain.ErrInvalidCredentials
	}
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (userRepository) List(ctx context.Context, db domain.DBTX, p domain.Pagination, search string) ([]*domain.User, int64, error) {
	var (
		where string
		args  []any
	)
	if search != "" {
		where = " WHERE username LIKE ? OR email LIKE ? OR full_name LIKE ?"
		like := "%" + search + "%"
		args = append(args, like, like, like)
	}
	total, err := countRows(ctx, db, `SELECT COUNT(*) FROM users`+where, args...)
	if err != nil {
		return nil, 0, err
	}
	q := `SELECT id, username, email, full_name, status, created_at, updated_at FROM users` + where +
		` ORDER BY id DESC LIMIT ? OFFSET ?`
	args = append(args, p.PageSize, p.Offset())
	rows, err := db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out := make([]*domain.User, 0)
	for rows.Next() {
		u := &domain.User{}
		if err := rows.Scan(&u.ID, &u.Username, &u.Email, &u.FullName, &u.Status, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, 0, err
		}
		out = append(out, u)
	}
	return out, total, rows.Err()
}

func (userRepository) Delete(ctx context.Context, db domain.DBTX, id int64) error {
	_, err := db.ExecContext(ctx, `DELETE FROM users WHERE id=?`, id)
	return fkOrErr(err, "存在关联数据，无法删除用户")
}

func (userRepository) AssignRoles(ctx context.Context, db domain.DBTX, userID int64, roleIDs []int64) error {
	if _, err := db.ExecContext(ctx, `DELETE FROM user_roles WHERE user_id=?`, userID); err != nil {
		return err
	}
	for _, rid := range roleIDs {
		if _, err := db.ExecContext(ctx, `INSERT INTO user_roles (user_id, role_id) VALUES (?, ?)`, userID, rid); err != nil {
			return fkOrErr(err, "角色不存在")
		}
	}
	return nil
}

func (userRepository) ListRoleIDs(ctx context.Context, db domain.DBTX, userID int64) ([]int64, error) {
	rows, err := db.QueryContext(ctx, `SELECT role_id FROM user_roles WHERE user_id=?`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]int64, 0)
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out = append(out, id)
	}
	return out, rows.Err()
}

func (userRepository) ListRoles(ctx context.Context, db domain.DBTX, userID int64) ([]*domain.Role, error) {
	q := `SELECT r.id, r.code, r.name, r.description, r.created_at, r.updated_at
	      FROM roles r JOIN user_roles ur ON ur.role_id = r.id WHERE ur.user_id=? ORDER BY r.id`
	rows, err := db.QueryContext(ctx, q, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]*domain.Role, 0)
	for rows.Next() {
		r := &domain.Role{}
		if err := rows.Scan(&r.ID, &r.Code, &r.Name, &r.Description, &r.CreatedAt, &r.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// scanUserFull scans all user columns including the password hash.
func scanUserFull(rows *sql.Rows, u *domain.User) error {
	return rows.Scan(&u.ID, &u.Username, &u.Email, &u.FullName, &u.PasswordHash,
		&u.Status, &u.CreatedAt, &u.UpdatedAt)
}

// markUserUpdated touches updated_at (used after role assignment).
func markUserUpdated(ctx context.Context, db domain.DBTX, id int64) {
	_, _ = db.ExecContext(ctx, `UPDATE users SET updated_at=? WHERE id=?`, time.Now(), id)
}
