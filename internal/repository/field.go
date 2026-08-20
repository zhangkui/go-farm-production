package repository

import (
	"context"
	"database/sql"

	"go-farm-production/internal/domain"
)

// FieldRepository persists fields.
type FieldRepository interface {
	Create(ctx context.Context, db domain.DBTX, f *domain.Field) (int64, error)
	Update(ctx context.Context, db domain.DBTX, id int64, u *domain.FieldUpsert) error
	GetByID(ctx context.Context, db domain.DBTX, id int64) (*domain.Field, error)
	List(ctx context.Context, db domain.DBTX, p domain.Pagination, farmID *int64, irrigationZone string) ([]*domain.Field, int64, error)
	Delete(ctx context.Context, db domain.DBTX, id int64) error
}

type fieldRepository struct{}

// NewFieldRepository returns the default MySQL FieldRepository.
func NewFieldRepository(db *sql.DB) FieldRepository { return fieldRepository{} }

func (fieldRepository) Create(ctx context.Context, db domain.DBTX, f *domain.Field) (int64, error) {
	q := `INSERT INTO fields (farm_id, code, name, area, soil_type, irrigation_zone, status, remark)
	      VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	id, err := execInsert(ctx, db, q, f.FarmID, f.Code, f.Name, f.Area, f.SoilType, f.IrrigationZone, f.Status, f.Remark)
	if err != nil {
		return 0, dupOrErr(err, "地块编码在该农场内已存在")
	}
	return id, nil
}

func (fieldRepository) Update(ctx context.Context, db domain.DBTX, id int64, u *domain.FieldUpsert) error {
	_, err := db.ExecContext(ctx,
		`UPDATE fields SET farm_id=?, code=?, name=?, area=?, soil_type=?, irrigation_zone=?, status=?, remark=? WHERE id=?`,
		u.FarmID, u.Code, u.Name, u.Area, u.SoilType, u.IrrigationZone, u.Status, u.Remark, id)
	return dupOrErr(err, "地块编码在该农场内已存在")
}

func (fieldRepository) GetByID(ctx context.Context, db domain.DBTX, id int64) (*domain.Field, error) {
	f := &domain.Field{}
	q := `SELECT f.id, f.farm_id, fa.name, f.code, f.name, f.area, f.soil_type, f.irrigation_zone, f.status, f.remark, f.created_at, f.updated_at
	      FROM fields f LEFT JOIN farms fa ON fa.id = f.farm_id WHERE f.id=?`
	err := db.QueryRowContext(ctx, q, id).Scan(&f.ID, &f.FarmID, &f.FarmName, &f.Code, &f.Name,
		&f.Area, &f.SoilType, &f.IrrigationZone, &f.Status, &f.Remark, &f.CreatedAt, &f.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, notFound("地块")
	}
	return f, err
}

func (fieldRepository) List(ctx context.Context, db domain.DBTX, p domain.Pagination, farmID *int64, irrigationZone string) ([]*domain.Field, int64, error) {
	var conds []string
	var args []any
	if farmID != nil {
		conds = append(conds, "f.farm_id=?")
		args = append(args, *farmID)
	}
	if irrigationZone != "" {
		conds = append(conds, "f.irrigation_zone=?")
		args = append(args, irrigationZone)
	}
	where := ""
	if len(conds) > 0 {
		where = " WHERE " + joinAnd(conds)
	}
	total, err := countRows(ctx, db, `SELECT COUNT(*) FROM fields f`+where, args...)
	if err != nil {
		return nil, 0, err
	}
	q := `SELECT f.id, f.farm_id, fa.name, f.code, f.name, f.area, f.soil_type, f.irrigation_zone, f.status, f.remark, f.created_at, f.updated_at
	      FROM fields f LEFT JOIN farms fa ON fa.id = f.farm_id` + where +
		` ORDER BY f.id DESC LIMIT ? OFFSET ?`
	args = append(args, p.PageSize, p.Offset())
	rows, err := db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out := make([]*domain.Field, 0)
	for rows.Next() {
		f := &domain.Field{}
		if err := rows.Scan(&f.ID, &f.FarmID, &f.FarmName, &f.Code, &f.Name,
			&f.Area, &f.SoilType, &f.IrrigationZone, &f.Status, &f.Remark, &f.CreatedAt, &f.UpdatedAt); err != nil {
			return nil, 0, err
		}
		out = append(out, f)
	}
	return out, total, rows.Err()
}

func (fieldRepository) Delete(ctx context.Context, db domain.DBTX, id int64) error {
	_, err := db.ExecContext(ctx, `DELETE FROM fields WHERE id=?`, id)
	return fkOrErr(err, "存在关联种植计划，无法删除地块")
}

// joinAnd joins conditions with AND.
func joinAnd(parts []string) string {
	out := ""
	for i, s := range parts {
		if i > 0 {
			out += " AND "
		}
		out += s
	}
	return out
}
