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
	CreateScoped(ctx context.Context, db domain.DBTX, t *domain.FarmTask, scope domain.TaskCreateScope) (int64, error)
	Update(ctx context.Context, db domain.DBTX, id int64, t *domain.FarmTask) error
	UpdateStatus(ctx context.Context, db domain.DBTX, id int64, status int8, completedDate *time.Time) error
	GetByID(ctx context.Context, db domain.DBTX, id int64) (*domain.FarmTask, error)
	List(ctx context.Context, db domain.DBTX, p domain.Pagination, filter TaskFilter) ([]*domain.FarmTask, int64, error)
	Delete(ctx context.Context, db domain.DBTX, id int64) error
	ListByPlan(ctx context.Context, db domain.DBTX, planID int64) ([]*domain.FarmTask, error)
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

func (r farmTaskRepository) CreateScoped(ctx context.Context, db domain.DBTX, t *domain.FarmTask, scope domain.TaskCreateScope) (int64, error) {
	return r.Create(ctx, db, t)
}

func (farmTaskRepository) Update(ctx context.Context, db domain.DBTX, id int64, t *domain.FarmTask) error {
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
