package repository

import (
	"context"
	"database/sql"
	"time"

	"go-farm-production/internal/domain"
)

// FarmTaskRepository persists farm tasks.
type FarmTaskRepository interface {
	Create(ctx context.Context, db domain.DBTX, t *domain.FarmTask) (int64, error)
	Update(ctx context.Context, db domain.DBTX, id int64, t *domain.FarmTask) error
	ValidateAccountingIdentity(ctx context.Context, db domain.DBTX, id int64, t *domain.FarmTask) error
	HasInputUsage(ctx context.Context, db domain.DBTX, id int64) (bool, error)
	UpdateStatus(ctx context.Context, db domain.DBTX, id int64, status int8, completedDate *time.Time) error
	GetByID(ctx context.Context, db domain.DBTX, id int64) (*domain.FarmTask, error)
	List(ctx context.Context, db domain.DBTX, p domain.Pagination, filter TaskFilter) ([]*domain.FarmTask, int64, error)
	Delete(ctx context.Context, db domain.DBTX, id int64) error
	ListByPlan(ctx context.Context, db domain.DBTX, planID int64) ([]*domain.FarmTask, error)
}

// ValidateAccountingIdentity enforces the immutability of a task's accounting
// identity (planting plan + task type) once the task has any input-material
// usage. Allocate (领用) and waste (损耗) both consume stock and are booked
// against the task's original plan and cost category, so re- attributing them
// to a different plan or task type would corrupt historical cost records.
//
// It is deliberately self-contained: given the requested task values it refuses
// any identity change when usage exists, so callers that invoke the repository
// directly (bypassing the service layer) are still protected.
func (farmTaskRepository) ValidateAccountingIdentity(ctx context.Context, db domain.DBTX, id int64, task *domain.FarmTask) error {
	var planID int64
	var taskType int8
	var allocations int
	err := db.QueryRowContext(ctx, `SELECT t.planting_plan_id, t.task_type,
		(SELECT COUNT(*) FROM input_allocations a
		 WHERE a.task_id = t.id AND a.type IN (?, ?))
		FROM farm_tasks t WHERE t.id=?`,
		domain.AllocationTypeAllocate, domain.AllocationTypeWaste, id).
		Scan(&planID, &taskType, &allocations)
	if err != nil {
		return err
	}
	if allocations > 0 && (planID != task.PlantingPlanID || taskType != task.TaskType) {
		return domain.Wrap(domain.CodeConflict, 409,
			"任务已产生投入品领用或损耗记录，不可变更所属种植计划或任务类型", nil)
	}
	return nil
}

// HasInputUsage reports whether a task has any input-material consumption
// (allocate or waste). Allocate and waste both deduct stock and are therefore
// committed usage; a return does not on its own count as usage (it only moves
// previously allocated stock back). Used by the service layer to refuse
// accounting-identity changes before building the write payload.
func (farmTaskRepository) HasInputUsage(ctx context.Context, db domain.DBTX, id int64) (bool, error) {
	var n int
	err := db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM input_allocations WHERE task_id=? AND type IN (?, ?)`,
		id, domain.AllocationTypeAllocate, domain.AllocationTypeWaste).Scan(&n)
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

// TaskFilter narrows a task listing query.
type TaskFilter struct {
	PlanID     *int64
	Status     *int8
	TaskType   *int8
	AssigneeID *int64
}

type farmTaskRepository struct{}

// NewFarmTaskRepository returns the default MySQL FarmTaskRepository.
func NewFarmTaskRepository(db *sql.DB) FarmTaskRepository { return farmTaskRepository{} }

func (farmTaskRepository) Create(ctx context.Context, db domain.DBTX, t *domain.FarmTask) (int64, error) {
	q := `INSERT INTO farm_tasks
	      (planting_plan_id, task_type, title, description, planned_date, completed_date, status,
	       labour_hours, equipment_cost, assignee_id, remark)
	      VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	id, err := execInsert(ctx, db, q, t.PlantingPlanID, t.TaskType, t.Title, t.Description,
		t.PlannedDate, t.CompletedDate, t.Status, t.LabourHours, t.EquipmentCost, t.AssigneeID, t.Remark)
	if err != nil {
		return 0, err
	}
	return id, nil
}

// Update writes a task. It first enforces ValidateAccountingIdentity so that
// the immutability invariant holds even when the repository is called directly,
// bypassing the service layer.
func (farmTaskRepository) Update(ctx context.Context, db domain.DBTX, id int64, t *domain.FarmTask) error {
	if err := (farmTaskRepository{}).ValidateAccountingIdentity(ctx, db, id, t); err != nil {
		return err
	}
	q := `UPDATE farm_tasks SET planting_plan_id=?, task_type=?, title=?, description=?, planned_date=?,
	      completed_date=?, status=?, labour_hours=?, equipment_cost=?, assignee_id=?, remark=? WHERE id=?`
	_, err := db.ExecContext(ctx, q, t.PlantingPlanID, t.TaskType, t.Title, t.Description,
		t.PlannedDate, t.CompletedDate, t.Status, t.LabourHours, t.EquipmentCost, t.AssigneeID, t.Remark, id)
	return err
}

func (farmTaskRepository) UpdateStatus(ctx context.Context, db domain.DBTX, id int64, status int8, completedDate *time.Time) error {
	if completedDate != nil {
		_, err := db.ExecContext(ctx, `UPDATE farm_tasks SET status=?, completed_date=? WHERE id=?`, status, completedDate, id)
		return err
	}
	_, err := db.ExecContext(ctx, `UPDATE farm_tasks SET status=? WHERE id=?`, status, id)
	return err
}

func (farmTaskRepository) GetByID(ctx context.Context, db domain.DBTX, id int64) (*domain.FarmTask, error) {
	q := `SELECT t.id, t.planting_plan_id, t.task_type, t.title, t.description, t.planned_date, t.completed_date,
	             t.status, t.labour_hours, t.equipment_cost, t.assignee_id, u.full_name, t.remark, t.created_at, t.updated_at
	      FROM farm_tasks t LEFT JOIN users u ON u.id = t.assignee_id WHERE t.id=?`
	t := &domain.FarmTask{}
	err := db.QueryRowContext(ctx, q, id).Scan(&t.ID, &t.PlantingPlanID, &t.TaskType, &t.Title,
		&t.Description, &t.PlannedDate, &t.CompletedDate, &t.Status, &t.LabourHours, &t.EquipmentCost,
		&t.AssigneeID, &t.AssigneeName, &t.Remark, &t.CreatedAt, &t.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, notFound("农事任务")
	}
	return t, err
}

func (farmTaskRepository) List(ctx context.Context, db domain.DBTX, p domain.Pagination, f TaskFilter) ([]*domain.FarmTask, int64, error) {
	var conds []string
	var args []any
	if f.PlanID != nil {
		conds = append(conds, "t.planting_plan_id=?")
		args = append(args, *f.PlanID)
	}
	if f.Status != nil {
		conds = append(conds, "t.status=?")
		args = append(args, *f.Status)
	}
	if f.TaskType != nil {
		conds = append(conds, "t.task_type=?")
		args = append(args, *f.TaskType)
	}
	if f.AssigneeID != nil {
		conds = append(conds, "t.assignee_id=?")
		args = append(args, *f.AssigneeID)
	}
	where := ""
	if len(conds) > 0 {
		where = " WHERE " + joinAnd(conds)
	}
	total, err := countRows(ctx, db, `SELECT COUNT(*) FROM farm_tasks t`+where, args...)
	if err != nil {
		return nil, 0, err
	}
	q := `SELECT t.id, t.planting_plan_id, t.task_type, t.title, COALESCE(t.description,''), t.planned_date, t.completed_date,
	             t.status, t.labour_hours, t.equipment_cost, t.assignee_id, COALESCE(u.full_name,''), COALESCE(t.remark,''), t.created_at, t.updated_at
	      FROM farm_tasks t LEFT JOIN users u ON u.id = t.assignee_id` + where +
		` ORDER BY t.id DESC LIMIT ? OFFSET ?`
	args = append(args, p.PageSize, p.Offset())
	rows, err := db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out := make([]*domain.FarmTask, 0)
	for rows.Next() {
		t := &domain.FarmTask{}
		if err := rows.Scan(&t.ID, &t.PlantingPlanID, &t.TaskType, &t.Title, &t.Description,
			&t.PlannedDate, &t.CompletedDate, &t.Status, &t.LabourHours, &t.EquipmentCost,
			&t.AssigneeID, &t.AssigneeName, &t.Remark, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, 0, err
		}
		out = append(out, t)
	}
	return out, total, rows.Err()
}

func (farmTaskRepository) ListByPlan(ctx context.Context, db domain.DBTX, planID int64) ([]*domain.FarmTask, error) {
	q := `SELECT t.id, t.planting_plan_id, t.task_type, t.title, COALESCE(t.description,''), t.planned_date, t.completed_date,
	             t.status, t.labour_hours, t.equipment_cost, t.assignee_id, COALESCE(u.full_name,''), COALESCE(t.remark,''), t.created_at, t.updated_at
	      FROM farm_tasks t LEFT JOIN users u ON u.id = t.assignee_id WHERE t.planting_plan_id=? ORDER BY t.planned_date`
	rows, err := db.QueryContext(ctx, q, planID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]*domain.FarmTask, 0)
	for rows.Next() {
		t := &domain.FarmTask{}
		if err := rows.Scan(&t.ID, &t.PlantingPlanID, &t.TaskType, &t.Title, &t.Description,
			&t.PlannedDate, &t.CompletedDate, &t.Status, &t.LabourHours, &t.EquipmentCost,
			&t.AssigneeID, &t.AssigneeName, &t.Remark, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

func (farmTaskRepository) Delete(ctx context.Context, db domain.DBTX, id int64) error {
	_, err := db.ExecContext(ctx, `DELETE FROM farm_tasks WHERE id=?`, id)
	return fkOrErr(err, "存在关联投入品领用，无法删除任务")
}
