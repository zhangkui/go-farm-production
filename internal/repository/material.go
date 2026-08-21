package repository

import (
	"context"
	"database/sql"

	"go-farm-production/internal/domain"
)

// MaterialRepository persists input materials.
type MaterialRepository interface {
	Create(ctx context.Context, db domain.DBTX, m *domain.Material) (int64, error)
	Update(ctx context.Context, db domain.DBTX, id int64, u *domain.MaterialUpsert) error
	GetByID(ctx context.Context, db domain.DBTX, id int64) (*domain.Material, error)
	List(ctx context.Context, db domain.DBTX, p domain.Pagination, category *int8) ([]*domain.Material, int64, error)
	Delete(ctx context.Context, db domain.DBTX, id int64) error
}

type materialRepository struct{}

// NewMaterialRepository returns the default MySQL MaterialRepository.
func NewMaterialRepository(db *sql.DB) MaterialRepository { return materialRepository{} }

func (materialRepository) Create(ctx context.Context, db domain.DBTX, m *domain.Material) (int64, error) {
	q := `INSERT INTO materials (code, name, category, unit, unit_price, description, status) VALUES (?, ?, ?, ?, ?, ?, ?)`
	id, err := execInsert(ctx, db, q, m.Code, m.Name, m.Category, m.Unit, m.UnitPrice, m.Description, m.Status)
	if err != nil {
		return 0, dupOrErr(err, "物料编码已存在")
	}
	return id, nil
}

func (materialRepository) Update(ctx context.Context, db domain.DBTX, id int64, u *domain.MaterialUpsert) error {
	_, err := db.ExecContext(ctx,
		`UPDATE materials SET code=?, name=?, category=?, unit=?, unit_price=?, description=?, status=? WHERE id=?`,
		u.Code, u.Name, u.Category, u.Unit, u.UnitPrice, u.Description, u.Status, id)
	return dupOrErr(err, "物料编码已存在")
}

func (materialRepository) GetByID(ctx context.Context, db domain.DBTX, id int64) (*domain.Material, error) {
	m := &domain.Material{}
	err := db.QueryRowContext(ctx,
		`SELECT id, code, name, category, unit, unit_price, description, status, created_at, updated_at FROM materials WHERE id=?`, id).
		Scan(&m.ID, &m.Code, &m.Name, &m.Category, &m.Unit, &m.UnitPrice, &m.Description, &m.Status, &m.CreatedAt, &m.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, notFound("物料")
	}
	return m, err
}

func (materialRepository) List(ctx context.Context, db domain.DBTX, p domain.Pagination, category *int8) ([]*domain.Material, int64, error) {
	var where string
	var args []any
	if category != nil {
		where = " WHERE category=?"
		args = append(args, *category)
	}
	total, err := countRows(ctx, db, `SELECT COUNT(*) FROM materials`+where, args...)
	if err != nil {
		return nil, 0, err
	}
	q := `SELECT id, code, name, category, unit, unit_price, description, status, created_at, updated_at FROM materials` + where +
		` ORDER BY id DESC LIMIT ? OFFSET ?`
	args = append(args, p.PageSize, p.Offset())
	rows, err := db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out := make([]*domain.Material, 0)
	for rows.Next() {
		m := &domain.Material{}
		if err := rows.Scan(&m.ID, &m.Code, &m.Name, &m.Category, &m.Unit, &m.UnitPrice, &m.Description, &m.Status, &m.CreatedAt, &m.UpdatedAt); err != nil {
			return nil, 0, err
		}
		out = append(out, m)
	}
	return out, total, rows.Err()
}

func (materialRepository) Delete(ctx context.Context, db domain.DBTX, id int64) error {
	_, err := db.ExecContext(ctx, `DELETE FROM materials WHERE id=?`, id)
	return fkOrErr(err, "存在关联批次，无法删除物料")
}
