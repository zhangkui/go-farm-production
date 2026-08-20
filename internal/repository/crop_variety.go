package repository

import (
	"context"
	"database/sql"

	"go-farm-production/internal/domain"
)

// CropVarietyRepository persists crop varieties.
type CropVarietyRepository interface {
	Create(ctx context.Context, db domain.DBTX, v *domain.CropVariety) (int64, error)
	Update(ctx context.Context, db domain.DBTX, id int64, u *domain.CropVarietyUpsert) error
	GetByID(ctx context.Context, db domain.DBTX, id int64) (*domain.CropVariety, error)
	List(ctx context.Context, db domain.DBTX, p domain.Pagination, category string) ([]*domain.CropVariety, int64, error)
	Delete(ctx context.Context, db domain.DBTX, id int64) error
	CountPlanReferences(ctx context.Context, db domain.DBTX, id int64) (int64, error)
	MarkDeleting(ctx context.Context, db domain.DBTX, id int64) error
	DeleteWithState(ctx context.Context, db domain.DBTX, state domain.VarietyDeletionState) error
}

type cropVarietyRepository struct{}

// NewCropVarietyRepository returns the default MySQL CropVarietyRepository.
func NewCropVarietyRepository(db *sql.DB) CropVarietyRepository { return cropVarietyRepository{} }

func (cropVarietyRepository) Create(ctx context.Context, db domain.DBTX, v *domain.CropVariety) (int64, error) {
	q := `INSERT INTO crop_varieties (code, name, category, growth_cycle, description, status) VALUES (?, ?, ?, ?, ?, ?)`
	id, err := execInsert(ctx, db, q, v.Code, v.Name, v.Category, v.GrowthCycle, v.Description, v.Status)
	if err != nil {
		return 0, dupOrErr(err, "品种编码已存在")
	}
	return id, nil
}

func (cropVarietyRepository) Update(ctx context.Context, db domain.DBTX, id int64, u *domain.CropVarietyUpsert) error {
	_, err := db.ExecContext(ctx,
		`UPDATE crop_varieties SET code=?, name=?, category=?, growth_cycle=?, description=?, status=? WHERE id=?`,
		u.Code, u.Name, u.Category, u.GrowthCycle, u.Description, u.Status, id)
	return dupOrErr(err, "品种编码已存在")
}

func (cropVarietyRepository) GetByID(ctx context.Context, db domain.DBTX, id int64) (*domain.CropVariety, error) {
	v := &domain.CropVariety{}
	err := db.QueryRowContext(ctx,
		`SELECT id, code, name, category, growth_cycle, description, status, created_at, updated_at FROM crop_varieties WHERE id=?`, id).
		Scan(&v.ID, &v.Code, &v.Name, &v.Category, &v.GrowthCycle, &v.Description, &v.Status, &v.CreatedAt, &v.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, notFound("作物品种")
	}
	return v, err
}

func (cropVarietyRepository) List(ctx context.Context, db domain.DBTX, p domain.Pagination, category string) ([]*domain.CropVariety, int64, error) {
	var where string
	var args []any
	if category != "" {
		where = " WHERE category=?"
		args = append(args, category)
	}
	total, err := countRows(ctx, db, `SELECT COUNT(*) FROM crop_varieties`+where, args...)
	if err != nil {
		return nil, 0, err
	}
	q := `SELECT id, code, name, category, growth_cycle, description, status, created_at, updated_at FROM crop_varieties` + where +
		` ORDER BY id DESC LIMIT ? OFFSET ?`
	args = append(args, p.PageSize, p.Offset())
	rows, err := db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out := make([]*domain.CropVariety, 0)
	for rows.Next() {
		v := &domain.CropVariety{}
		if err := rows.Scan(&v.ID, &v.Code, &v.Name, &v.Category, &v.GrowthCycle, &v.Description, &v.Status, &v.CreatedAt, &v.UpdatedAt); err != nil {
			return nil, 0, err
		}
		out = append(out, v)
	}
	return out, total, rows.Err()
}

func (cropVarietyRepository) Delete(ctx context.Context, db domain.DBTX, id int64) error {
	_, err := db.ExecContext(ctx, `DELETE FROM crop_varieties WHERE id=?`, id)
	return fkOrErr(err, "存在关联种植计划，无法删除品种")
}

func (cropVarietyRepository) CountPlanReferences(ctx context.Context, db domain.DBTX, id int64) (int64, error) {
	return countRows(ctx, db, `SELECT COUNT(*) FROM planting_plans WHERE crop_variety_id=?`, id)
}

func (cropVarietyRepository) MarkDeleting(ctx context.Context, db domain.DBTX, id int64) error {
	_, err := db.ExecContext(ctx, `UPDATE crop_varieties SET status=? WHERE id=?`, domain.StatusInactive, id)
	return err
}

func (cropVarietyRepository) DeleteWithState(ctx context.Context, db domain.DBTX, state domain.VarietyDeletionState) error {
	_, err := db.ExecContext(ctx, `DELETE FROM crop_varieties WHERE id=?`, state.VarietyID)
	if isFKConstraint(err) {
		return domain.Wrap(state.ConflictCode(), 409, "品种仍被业务数据引用", err)
	}
	return err
}
