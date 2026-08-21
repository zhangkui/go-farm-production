package repository

import (
	"context"
	"database/sql"

	"go-farm-production/internal/domain"
)

// SeasonRepository persists planting seasons.
type SeasonRepository interface {
	Create(ctx context.Context, db domain.DBTX, s *domain.Season) (int64, error)
	Update(ctx context.Context, db domain.DBTX, id int64, u *domain.SeasonUpsert) error
	GetByID(ctx context.Context, db domain.DBTX, id int64) (*domain.Season, error)
	List(ctx context.Context, db domain.DBTX, p domain.Pagination, status *int8) ([]*domain.Season, int64, error)
	Delete(ctx context.Context, db domain.DBTX, id int64) error
}

type seasonRepository struct{}

// NewSeasonRepository returns the default MySQL SeasonRepository.
func NewSeasonRepository(db *sql.DB) SeasonRepository { return seasonRepository{} }

func (seasonRepository) Create(ctx context.Context, db domain.DBTX, s *domain.Season) (int64, error) {
	q := `INSERT INTO seasons (code, name, start_date, end_date, status) VALUES (?, ?, ?, ?, ?)`
	id, err := execInsert(ctx, db, q, s.Code, s.Name, s.StartDate, s.EndDate, s.Status)
	if err != nil {
		return 0, dupOrErr(err, "季次编码已存在")
	}
	return id, nil
}

func (seasonRepository) Update(ctx context.Context, db domain.DBTX, id int64, u *domain.SeasonUpsert) error {
	_, err := db.ExecContext(ctx,
		`UPDATE seasons SET code=?, name=?, start_date=?, end_date=?, status=? WHERE id=?`,
		u.Code, u.Name, u.StartDate, u.EndDate, u.Status, id)
	return dupOrErr(err, "季次编码已存在")
}

func (seasonRepository) GetByID(ctx context.Context, db domain.DBTX, id int64) (*domain.Season, error) {
	s := &domain.Season{}
	err := db.QueryRowContext(ctx,
		`SELECT id, code, name, start_date, end_date, status, created_at, updated_at FROM seasons WHERE id=?`, id).
		Scan(&s.ID, &s.Code, &s.Name, &s.StartDate, &s.EndDate, &s.Status, &s.CreatedAt, &s.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, notFound("种植季")
	}
	return s, err
}

func (seasonRepository) List(ctx context.Context, db domain.DBTX, p domain.Pagination, status *int8) ([]*domain.Season, int64, error) {
	var where string
	var args []any
	if status != nil {
		where = " WHERE status=?"
		args = append(args, *status)
	}
	total, err := countRows(ctx, db, `SELECT COUNT(*) FROM seasons`+where, args...)
	if err != nil {
		return nil, 0, err
	}
	q := `SELECT id, code, name, start_date, end_date, status, created_at, updated_at FROM seasons` + where +
		` ORDER BY start_date DESC LIMIT ? OFFSET ?`
	args = append(args, p.PageSize, p.Offset())
	rows, err := db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out := make([]*domain.Season, 0)
	for rows.Next() {
		s := &domain.Season{}
		if err := rows.Scan(&s.ID, &s.Code, &s.Name, &s.StartDate, &s.EndDate, &s.Status, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, 0, err
		}
		out = append(out, s)
	}
	return out, total, rows.Err()
}

func (seasonRepository) Delete(ctx context.Context, db domain.DBTX, id int64) error {
	_, err := db.ExecContext(ctx, `DELETE FROM seasons WHERE id=?`, id)
	return fkOrErr(err, "存在关联种植计划，无法删除季次")
}
