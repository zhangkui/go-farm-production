package repository

import (
	"context"
	"database/sql"

	"go-farm-production/internal/domain"
)

// HarvestRepository persists harvest events.
type HarvestRepository interface {
	Create(ctx context.Context, db domain.DBTX, h *domain.Harvest) (int64, error)
	Update(ctx context.Context, db domain.DBTX, id int64, h *domain.Harvest) error
	SetApproved(ctx context.Context, db domain.DBTX, id int64, approved bool) error
	SetApprovedForPlan(ctx context.Context, db domain.DBTX, planID int64, approved bool) error
	HasApprovedForPlan(ctx context.Context, db domain.DBTX, planID int64) (bool, error)
	GetByID(ctx context.Context, db domain.DBTX, id int64) (*domain.Harvest, error)
	List(ctx context.Context, db domain.DBTX, p domain.Pagination, filter HarvestFilter) ([]*domain.Harvest, int64, error)
	Delete(ctx context.Context, db domain.DBTX, id int64) error
	ListByPlan(ctx context.Context, db domain.DBTX, planID int64) ([]*domain.Harvest, error)
}

// HarvestFilter narrows a harvest listing query.
type HarvestFilter struct {
	PlanID   *int64
	Approved *bool
}

type harvestRepository struct{}

// NewHarvestRepository returns the default MySQL HarvestRepository.
func NewHarvestRepository(db *sql.DB) HarvestRepository { return harvestRepository{} }

func (harvestRepository) Create(ctx context.Context, db domain.DBTX, h *domain.Harvest) (int64, error) {
	q := `INSERT INTO harvests (planting_plan_id, harvest_date, total_weight, grade, remark, approved)
	      VALUES (?, ?, ?, ?, ?, ?)`
	return execInsert(ctx, db, q, h.PlantingPlanID, h.HarvestDate, h.TotalWeight, h.Grade, h.Remark, h.Approved)
}

func (harvestRepository) Update(ctx context.Context, db domain.DBTX, id int64, h *domain.Harvest) error {
	q := `UPDATE harvests SET planting_plan_id=?, harvest_date=?, total_weight=?, grade=?, remark=?, approved=? WHERE id=?`
	_, err := db.ExecContext(ctx, q, h.PlantingPlanID, h.HarvestDate, h.TotalWeight, h.Grade, h.Remark, h.Approved, id)
	return err
}

func (harvestRepository) SetApproved(ctx context.Context, db domain.DBTX, id int64, approved bool) error {
	_, err := db.ExecContext(ctx, `UPDATE harvests SET approved=? WHERE id=?`, approved, id)
	return err
}

func (harvestRepository) SetApprovedForPlan(ctx context.Context, db domain.DBTX, planID int64, approved bool) error {
	_, err := db.ExecContext(ctx,
		`UPDATE harvests SET approved=? WHERE planting_plan_id=?`,
		approved, planID)
	return err
}

func (harvestRepository) HasApprovedForPlan(ctx context.Context, db domain.DBTX, planID int64) (bool, error) {
	var count int64
	err := db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM harvests WHERE planting_plan_id=? AND approved=1`,
		planID).Scan(&count)
	return count > 0, err
}

func (harvestRepository) GetByID(ctx context.Context, db domain.DBTX, id int64) (*domain.Harvest, error) {
	h := &domain.Harvest{}
	err := db.QueryRowContext(ctx,
		`SELECT id, planting_plan_id, harvest_date, total_weight, grade, remark, approved, created_at, updated_at FROM harvests WHERE id=?`, id).
		Scan(&h.ID, &h.PlantingPlanID, &h.HarvestDate, &h.TotalWeight, &h.Grade, &h.Remark, &h.Approved, &h.CreatedAt, &h.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, notFound("采收记录")
	}
	return h, err
}

func (harvestRepository) List(ctx context.Context, db domain.DBTX, p domain.Pagination, f HarvestFilter) ([]*domain.Harvest, int64, error) {
	var conds []string
	var args []any
	if f.PlanID != nil {
		conds = append(conds, "planting_plan_id=?")
		args = append(args, *f.PlanID)
	}
	if f.Approved != nil {
		conds = append(conds, "approved=?")
		args = append(args, *f.Approved)
	}
	where := ""
	if len(conds) > 0 {
		where = " WHERE " + joinAnd(conds)
	}
	total, err := countRows(ctx, db, `SELECT COUNT(*) FROM harvests`+where, args...)
	if err != nil {
		return nil, 0, err
	}
	q := `SELECT id, planting_plan_id, harvest_date, total_weight, grade, remark, approved, created_at, updated_at FROM harvests` + where +
		` ORDER BY id DESC LIMIT ? OFFSET ?`
	args = append(args, p.PageSize, p.Offset())
	rows, err := db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out := make([]*domain.Harvest, 0)
	for rows.Next() {
		h := &domain.Harvest{}
		if err := rows.Scan(&h.ID, &h.PlantingPlanID, &h.HarvestDate, &h.TotalWeight, &h.Grade, &h.Remark, &h.Approved, &h.CreatedAt, &h.UpdatedAt); err != nil {
			return nil, 0, err
		}
		out = append(out, h)
	}
	return out, total, rows.Err()
}

func (harvestRepository) ListByPlan(ctx context.Context, db domain.DBTX, planID int64) ([]*domain.Harvest, error) {
	q := `SELECT id, planting_plan_id, harvest_date, total_weight, grade, remark, approved, created_at, updated_at FROM harvests WHERE planting_plan_id=? ORDER BY harvest_date`
	rows, err := db.QueryContext(ctx, q, planID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]*domain.Harvest, 0)
	for rows.Next() {
		h := &domain.Harvest{}
		if err := rows.Scan(&h.ID, &h.PlantingPlanID, &h.HarvestDate, &h.TotalWeight, &h.Grade, &h.Remark, &h.Approved, &h.CreatedAt, &h.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, h)
	}
	return out, rows.Err()
}

func (harvestRepository) Delete(ctx context.Context, db domain.DBTX, id int64) error {
	_, err := db.ExecContext(ctx, `DELETE FROM harvests WHERE id=?`, id)
	return fkOrErr(err, "存在关联农产品库存，无法删除采收记录")
}
