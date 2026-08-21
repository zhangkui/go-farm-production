package repository

import (
	"context"
	"database/sql"

	"go-farm-production/internal/domain"
)

// CostAnalysisRepository persists cost analyses and supports rollup queries.
type CostAnalysisRepository interface {
	Create(ctx context.Context, db domain.DBTX, c *domain.CostAnalysis) (int64, error)
	Update(ctx context.Context, db domain.DBTX, id int64, c *domain.CostAnalysis) error
	GetByID(ctx context.Context, db domain.DBTX, id int64) (*domain.CostAnalysis, error)
	GetByPlan(ctx context.Context, db domain.DBTX, planID int64) (*domain.CostAnalysis, error)
	List(ctx context.Context, db domain.DBTX, p domain.Pagination, planID *int64) ([]*domain.CostAnalysis, int64, error)
	Delete(ctx context.Context, db domain.DBTX, id int64) error
	// Summarise aggregates cost across plans for reporting, optionally filtered
	// by season id and/or crop variety id.
	Summarise(ctx context.Context, db domain.DBTX, seasonID, varietyID *int64) ([]*domain.CostSummary, error)
}

type costAnalysisRepository struct{}

// NewCostAnalysisRepository returns the default MySQL CostAnalysisRepository.
func NewCostAnalysisRepository(db *sql.DB) CostAnalysisRepository { return costAnalysisRepository{} }

func (costAnalysisRepository) Create(ctx context.Context, db domain.DBTX, c *domain.CostAnalysis) (int64, error) {
	q := `INSERT INTO cost_analyses (planting_plan_id, analysis_date, labour_cost, input_cost, equipment_cost,
	      other_cost, total_cost, total_yield, unit_cost, remark) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	return execInsert(ctx, db, q, c.PlantingPlanID, c.AnalysisDate, c.LabourCost, c.InputCost,
		c.EquipmentCost, c.OtherCost, c.TotalCost, c.TotalYield, c.UnitCost, c.Remark)
}

func (costAnalysisRepository) Update(ctx context.Context, db domain.DBTX, id int64, c *domain.CostAnalysis) error {
	q := `UPDATE cost_analyses SET planting_plan_id=?, analysis_date=?, labour_cost=?, input_cost=?, equipment_cost=?,
	      other_cost=?, total_cost=?, total_yield=?, unit_cost=?, remark=? WHERE id=?`
	_, err := db.ExecContext(ctx, q, c.PlantingPlanID, c.AnalysisDate, c.LabourCost, c.InputCost,
		c.EquipmentCost, c.OtherCost, c.TotalCost, c.TotalYield, c.UnitCost, c.Remark, id)
	return err
}

func (costAnalysisRepository) GetByID(ctx context.Context, db domain.DBTX, id int64) (*domain.CostAnalysis, error) {
	c := &domain.CostAnalysis{}
	err := db.QueryRowContext(ctx,
		`SELECT id, planting_plan_id, analysis_date, labour_cost, input_cost, equipment_cost, other_cost, total_cost, total_yield, unit_cost, remark, created_at, updated_at FROM cost_analyses WHERE id=?`, id).
		Scan(&c.ID, &c.PlantingPlanID, &c.AnalysisDate, &c.LabourCost, &c.InputCost, &c.EquipmentCost,
			&c.OtherCost, &c.TotalCost, &c.TotalYield, &c.UnitCost, &c.Remark, &c.CreatedAt, &c.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, notFound("成本分析")
	}
	return c, err
}

func (costAnalysisRepository) GetByPlan(ctx context.Context, db domain.DBTX, planID int64) (*domain.CostAnalysis, error) {
	c := &domain.CostAnalysis{}
	err := db.QueryRowContext(ctx,
		`SELECT id, planting_plan_id, analysis_date, labour_cost, input_cost, equipment_cost, other_cost, total_cost, total_yield, unit_cost, remark, created_at, updated_at FROM cost_analyses WHERE planting_plan_id=? ORDER BY id DESC LIMIT 1`, planID).
		Scan(&c.ID, &c.PlantingPlanID, &c.AnalysisDate, &c.LabourCost, &c.InputCost, &c.EquipmentCost,
			&c.OtherCost, &c.TotalCost, &c.TotalYield, &c.UnitCost, &c.Remark, &c.CreatedAt, &c.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, notFound("成本分析")
	}
	return c, err
}

func (costAnalysisRepository) List(ctx context.Context, db domain.DBTX, p domain.Pagination, planID *int64) ([]*domain.CostAnalysis, int64, error) {
	var where string
	var args []any
	if planID != nil {
		where = " WHERE planting_plan_id=?"
		args = append(args, *planID)
	}
	total, err := countRows(ctx, db, `SELECT COUNT(*) FROM cost_analyses`+where, args...)
	if err != nil {
		return nil, 0, err
	}
	q := `SELECT id, planting_plan_id, analysis_date, labour_cost, input_cost, equipment_cost, other_cost, total_cost, total_yield, unit_cost, remark, created_at, updated_at FROM cost_analyses` + where +
		` ORDER BY id DESC LIMIT ? OFFSET ?`
	args = append(args, p.PageSize, p.Offset())
	rows, err := db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out := make([]*domain.CostAnalysis, 0)
	for rows.Next() {
		c := &domain.CostAnalysis{}
		if err := rows.Scan(&c.ID, &c.PlantingPlanID, &c.AnalysisDate, &c.LabourCost, &c.InputCost,
			&c.EquipmentCost, &c.OtherCost, &c.TotalCost, &c.TotalYield, &c.UnitCost, &c.Remark,
			&c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, 0, err
		}
		out = append(out, c)
	}
	return out, total, rows.Err()
}

func (costAnalysisRepository) Delete(ctx context.Context, db domain.DBTX, id int64) error {
	_, err := db.ExecContext(ctx, `DELETE FROM cost_analyses WHERE id=?`, id)
	return err
}

func (costAnalysisRepository) Summarise(ctx context.Context, db domain.DBTX, seasonID, varietyID *int64) ([]*domain.CostSummary, error) {
	var conds []string
	var args []any
	if seasonID != nil {
		conds = append(conds, "p.season_id=?")
		args = append(args, *seasonID)
	}
	if varietyID != nil {
		conds = append(conds, "p.crop_variety_id=?")
		args = append(args, *varietyID)
	}
	where := ""
	if len(conds) > 0 {
		where = " WHERE " + joinAnd(conds)
	}
	q := `SELECT p.id, f.name, v.name, s.name,
	             COALESCE(c.total_cost,0), COALESCE(c.total_yield,0), COALESCE(c.unit_cost,0)
	      FROM planting_plans p
	      LEFT JOIN cost_analyses c ON c.planting_plan_id = p.id
	      LEFT JOIN fields f ON f.id = p.field_id
	      LEFT JOIN crop_varieties v ON v.id = p.crop_variety_id
	      LEFT JOIN seasons s ON s.id = p.season_id` + where +
		` ORDER BY COALESCE(c.unit_cost,0) DESC`
	rows, err := db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]*domain.CostSummary, 0)
	for rows.Next() {
		s := &domain.CostSummary{}
		if err := rows.Scan(&s.PlantingPlanID, &s.FieldName, &s.CropVarietyName, &s.SeasonName,
			&s.TotalCost, &s.TotalYield, &s.UnitCost); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}
