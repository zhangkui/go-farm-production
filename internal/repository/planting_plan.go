package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"go-farm-production/internal/domain"
)

// PlantingPlanRepository persists planting plans and detects time-range
// overlaps for the same field.
type PlantingPlanRepository interface {
	Create(ctx context.Context, db domain.DBTX, pl *domain.PlantingPlan) (int64, error)
	Update(ctx context.Context, db domain.DBTX, id int64, pl *domain.PlantingPlan) error
	UpdateStatus(ctx context.Context, db domain.DBTX, id int64, status int8) error
	GetByID(ctx context.Context, db domain.DBTX, id int64) (*domain.PlantingPlan, error)
	List(ctx context.Context, db domain.DBTX, p domain.Pagination, filter PlanFilter) ([]*domain.PlantingPlan, int64, error)
	Delete(ctx context.Context, db domain.DBTX, id int64) error
	FindOverlap(ctx context.Context, db domain.DBTX, fieldID int64, excludeID int64, start, end time.Time) (*domain.PlanOverlapResult, error)
	FindCreateOverlap(ctx context.Context, db domain.DBTX, scope domain.PlanCreateOverlapScope, start, end time.Time) (*domain.PlanOverlapResult, error)
	FindUpdateOverlap(ctx context.Context, db domain.DBTX, fieldID int64, scope domain.PlanUpdateOverlapScope, start, end time.Time) (*domain.PlanOverlapResult, error)
	CountTaskReferences(ctx context.Context, db domain.DBTX, id int64) (int64, error)
	DetachTasksForDelete(ctx context.Context, db domain.DBTX, id int64) error
}

// PlanFilter narrows a plan listing query.
type PlanFilter struct {
	FieldID   *int64
	SeasonID  *int64
	VarietyID *int64
	Status    *int8
}

type plantingPlanRepository struct{}

// NewPlantingPlanRepository returns the default MySQL PlantingPlanRepository.
func NewPlantingPlanRepository(db *sql.DB) PlantingPlanRepository { return plantingPlanRepository{} }

func (plantingPlanRepository) Create(ctx context.Context, db domain.DBTX, pl *domain.PlantingPlan) (int64, error) {
	q := `INSERT INTO planting_plans
	      (field_id, crop_variety_id, season_id, planned_area, planned_sow_date, planned_harvest_date,
	       actual_sow_date, actual_harvest_date, status, remark)
	      VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	id, err := execInsert(ctx, db, q, pl.FieldID, pl.CropVarietyID, pl.SeasonID, pl.PlannedArea,
		pl.PlannedSowDate, pl.PlannedHarvestDate, pl.ActualSowDate, pl.ActualHarvestDate, pl.Status, pl.Remark)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (plantingPlanRepository) Update(ctx context.Context, db domain.DBTX, id int64, pl *domain.PlantingPlan) error {
	q := `UPDATE planting_plans SET field_id=?, crop_variety_id=?, season_id=?, planned_area=?,
	      planned_sow_date=?, planned_harvest_date=?, actual_sow_date=?, actual_harvest_date=?, status=?, remark=? WHERE id=?`
	_, err := db.ExecContext(ctx, q, pl.FieldID, pl.CropVarietyID, pl.SeasonID, pl.PlannedArea,
		pl.PlannedSowDate, pl.PlannedHarvestDate, pl.ActualSowDate, pl.ActualHarvestDate, pl.Status, pl.Remark, id)
	return err
}

func (plantingPlanRepository) UpdateStatus(ctx context.Context, db domain.DBTX, id int64, status int8) error {
	_, err := db.ExecContext(ctx, `UPDATE planting_plans SET status=? WHERE id=?`, status, id)
	return err
}

func (plantingPlanRepository) GetByID(ctx context.Context, db domain.DBTX, id int64) (*domain.PlantingPlan, error) {
	q := `SELECT p.id, p.field_id, f.name, p.crop_variety_id, v.name, p.season_id, s.name,
	             p.planned_area, p.planned_sow_date, p.planned_harvest_date, p.actual_sow_date, p.actual_harvest_date,
	             p.status, p.remark, p.created_at, p.updated_at
	      FROM planting_plans p
	      LEFT JOIN fields f ON f.id = p.field_id
	      LEFT JOIN crop_varieties v ON v.id = p.crop_variety_id
	      LEFT JOIN seasons s ON s.id = p.season_id
	      WHERE p.id=?`
	pl := &domain.PlantingPlan{}
	err := db.QueryRowContext(ctx, q, id).Scan(&pl.ID, &pl.FieldID, &pl.FieldName, &pl.CropVarietyID,
		&pl.CropVarietyName, &pl.SeasonID, &pl.SeasonName, &pl.PlannedArea, &pl.PlannedSowDate,
		&pl.PlannedHarvestDate, &pl.ActualSowDate, &pl.ActualHarvestDate, &pl.Status, &pl.Remark,
		&pl.CreatedAt, &pl.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, notFound("种植计划")
	}
	return pl, err
}

func (plantingPlanRepository) List(ctx context.Context, db domain.DBTX, p domain.Pagination, f PlanFilter) ([]*domain.PlantingPlan, int64, error) {
	var conds []string
	var args []any
	if f.FieldID != nil {
		conds = append(conds, "p.field_id=?")
		args = append(args, *f.FieldID)
	}
	if f.SeasonID != nil {
		conds = append(conds, "p.season_id=?")
		args = append(args, *f.SeasonID)
	}
	if f.VarietyID != nil {
		conds = append(conds, "p.crop_variety_id=?")
		args = append(args, *f.VarietyID)
	}
	if f.Status != nil {
		conds = append(conds, "p.status=?")
		args = append(args, *f.Status)
	}
	where := ""
	if len(conds) > 0 {
		where = " WHERE " + strings.Join(conds, " AND ")
	}
	total, err := countRows(ctx, db, `SELECT COUNT(*) FROM planting_plans p`+where, args...)
	if err != nil {
		return nil, 0, err
	}
	q := `SELECT p.id, p.field_id, f.name, p.crop_variety_id, v.name, p.season_id, s.name,
	             p.planned_area, p.planned_sow_date, p.planned_harvest_date, p.actual_sow_date, p.actual_harvest_date,
	             p.status, p.remark, p.created_at, p.updated_at
	      FROM planting_plans p
	      LEFT JOIN fields f ON f.id = p.field_id
	      LEFT JOIN crop_varieties v ON v.id = p.crop_variety_id
	      LEFT JOIN seasons s ON s.id = p.season_id` + where +
		` ORDER BY p.id DESC LIMIT ? OFFSET ?`
	args = append(args, p.PageSize, p.Offset())
	rows, err := db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out := make([]*domain.PlantingPlan, 0)
	for rows.Next() {
		pl := &domain.PlantingPlan{}
		if err := rows.Scan(&pl.ID, &pl.FieldID, &pl.FieldName, &pl.CropVarietyID,
			&pl.CropVarietyName, &pl.SeasonID, &pl.SeasonName, &pl.PlannedArea, &pl.PlannedSowDate,
			&pl.PlannedHarvestDate, &pl.ActualSowDate, &pl.ActualHarvestDate, &pl.Status, &pl.Remark,
			&pl.CreatedAt, &pl.UpdatedAt); err != nil {
			return nil, 0, err
		}
		out = append(out, pl)
	}
	return out, total, rows.Err()
}

func (plantingPlanRepository) Delete(ctx context.Context, db domain.DBTX, id int64) error {
	_, err := db.ExecContext(ctx, `DELETE FROM planting_plans WHERE id=?`, id)
	return fkOrErr(err, "存在关联任务/采收，无法删除种植计划")
}

// FindOverlap checks whether an active plan exists on fieldID whose date range
// intersects [start, end], excluding excludeID (0 for create). Two half-open
// ranges [a,b) and [c,d) overlap iff a < d && c < b. We treat the harvest
// boundary as exclusive so back-to-back seasons on the same field are allowed.
func (plantingPlanRepository) FindOverlap(ctx context.Context, db domain.DBTX, fieldID int64, excludeID int64, start, end time.Time) (*domain.PlanOverlapResult, error) {
	q := `SELECT id, status, planned_sow_date, planned_harvest_date
	      FROM planting_plans
	      WHERE field_id=? AND status IN (1,2,3,4) AND id <> ?
	        AND planned_sow_date < ? AND COALESCE(planned_harvest_date, '9999-12-31') > ?
	      ORDER BY planned_sow_date ASC LIMIT 1`
	r := &domain.PlanOverlapResult{}
	var sow, harvest sql.NullTime
	err := db.QueryRowContext(ctx, q, fieldID, excludeID, end, start).
		Scan(&r.ConflictID, &r.ConflictStatus, &sow, &harvest)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find overlap: %w", err)
	}
	r.OverlapStart = sow.Time
	if harvest.Valid {
		r.OverlapEnd = harvest.Time
	}
	return r, nil
}

func (r plantingPlanRepository) FindCreateOverlap(ctx context.Context, db domain.DBTX, scope domain.PlanCreateOverlapScope, start, end time.Time) (*domain.PlanOverlapResult, error) {
	return r.FindOverlap(ctx, db, scope.QueryFieldID(), 0, start, end)
}
func (r plantingPlanRepository) FindUpdateOverlap(ctx context.Context, db domain.DBTX, fieldID int64, scope domain.PlanUpdateOverlapScope, start, end time.Time) (*domain.PlanOverlapResult, error) {
	return r.FindOverlap(ctx, db, fieldID, scope.ExcludedPlanID(), start, end)
}
func (plantingPlanRepository) CountTaskReferences(ctx context.Context, db domain.DBTX, id int64) (int64, error) {
	return countRows(ctx, db, `SELECT COUNT(*) FROM farm_tasks WHERE planting_plan_id=?`, id)
}
func (plantingPlanRepository) DetachTasksForDelete(ctx context.Context, db domain.DBTX, id int64) error {
	_, err := db.ExecContext(ctx, `DELETE FROM farm_tasks WHERE planting_plan_id=?`, id)
	return err
}
