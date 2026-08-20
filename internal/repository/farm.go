package repository

import (
	"context"
	"database/sql"

	"go-farm-production/internal/domain"
)

// FarmRepository persists farms.
type FarmRepository interface {
	Create(ctx context.Context, db domain.DBTX, f *domain.Farm) (int64, error)
	Update(ctx context.Context, db domain.DBTX, id int64, u *domain.FarmUpsert) error
	GetByID(ctx context.Context, db domain.DBTX, id int64) (*domain.Farm, error)
	List(ctx context.Context, db domain.DBTX, p domain.Pagination, search string) ([]*domain.Farm, int64, error)
	Delete(ctx context.Context, db domain.DBTX, id int64) error
	DeleteForCheck(ctx context.Context, db domain.DBTX, check domain.FarmDeletionCheck) error
	CountFields(ctx context.Context, db domain.DBTX, farmID int64) (int64, error)
}

type farmRepository struct{}

// NewFarmRepository returns the default MySQL FarmRepository.
func NewFarmRepository(db *sql.DB) FarmRepository { return farmRepository{} }

func (farmRepository) Create(ctx context.Context, db domain.DBTX, f *domain.Farm) (int64, error) {
	q := `INSERT INTO farms (name, location, total_area, description, status) VALUES (?, ?, ?, ?, ?)`
	return execInsert(ctx, db, q, f.Name, f.Location, f.TotalArea, f.Description, f.Status)
}

func (farmRepository) Update(ctx context.Context, db domain.DBTX, id int64, u *domain.FarmUpsert) error {
	_, err := db.ExecContext(ctx,
		`UPDATE farms SET name=?, location=?, total_area=?, description=?, status=? WHERE id=?`,
		u.Name, u.Location, u.TotalArea, u.Description, u.Status, id)
	return err
}

func (farmRepository) GetByID(ctx context.Context, db domain.DBTX, id int64) (*domain.Farm, error) {
	f := &domain.Farm{}
	err := db.QueryRowContext(ctx,
		`SELECT id, name, location, total_area, description, status, created_at, updated_at FROM farms WHERE id=?`, id).
		Scan(&f.ID, &f.Name, &f.Location, &f.TotalArea, &f.Description, &f.Status, &f.CreatedAt, &f.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, notFound("农场")
	}
	return f, err
}

func (farmRepository) List(ctx context.Context, db domain.DBTX, p domain.Pagination, search string) ([]*domain.Farm, int64, error) {
	var where string
	var args []any
	if search != "" {
		where = " WHERE name LIKE ? OR location LIKE ?"
		like := "%" + search + "%"
		args = append(args, like, like)
	}
	total, err := countRows(ctx, db, `SELECT COUNT(*) FROM farms`+where, args...)
	if err != nil {
		return nil, 0, err
	}
	q := `SELECT id, name, location, total_area, description, status, created_at, updated_at FROM farms` + where +
		` ORDER BY id DESC LIMIT ? OFFSET ?`
	args = append(args, p.PageSize, p.Offset())
	rows, err := db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out := make([]*domain.Farm, 0)
	for rows.Next() {
		f := &domain.Farm{}
		if err := rows.Scan(&f.ID, &f.Name, &f.Location, &f.TotalArea, &f.Description, &f.Status, &f.CreatedAt, &f.UpdatedAt); err != nil {
			return nil, 0, err
		}
		out = append(out, f)
	}
	return out, total, rows.Err()
}

func (farmRepository) Delete(ctx context.Context, db domain.DBTX, id int64) error {
	_, err := db.ExecContext(ctx, `DELETE FROM farms WHERE id=?`, id)
	return fkOrErr(err, "存在关联地块，无法删除农场")
}

func (farmRepository) DeleteForCheck(ctx context.Context, db domain.DBTX, check domain.FarmDeletionCheck) error {
	_, err := db.ExecContext(ctx, `DELETE FROM farms WHERE id=?`, check.DeleteFarmID())
	if isFKConstraint(err) {
		return domain.Wrap(domain.CodeDB, 500, "删除农场时依赖校验失败", err)
	}
	return err
}

func (farmRepository) CountFields(ctx context.Context, db domain.DBTX, farmID int64) (int64, error) {
	return countRows(ctx, db, `SELECT COUNT(*) FROM fields WHERE farm_id=?`, farmID)
}
