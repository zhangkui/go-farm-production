package repository

import (
	"context"
	"database/sql"

	"go-farm-production/internal/domain"
)

// InputAllocationRepository persists material movements (allocate/return/waste)
// against a batch. Stock mutation happens in the batch repo within the same tx.
type InputAllocationRepository interface {
	Create(ctx context.Context, db domain.DBTX, a *domain.InputAllocation) (int64, error)
	GetByID(ctx context.Context, db domain.DBTX, id int64) (*domain.InputAllocation, error)
	List(ctx context.Context, db domain.DBTX, p domain.Pagination, filter AllocFilter) ([]*domain.InputAllocation, int64, error)
	ListByTask(ctx context.Context, db domain.DBTX, taskID int64) ([]*domain.InputAllocation, error)
	ListByPlan(ctx context.Context, db domain.DBTX, planID int64) ([]*domain.InputAllocation, error)
	ReturnableQuantity(ctx context.Context, db domain.DBTX, batchID int64, taskID *int64) (domain.Decimal, error)
}

// AllocFilter narrows an allocation listing query.
type AllocFilter struct {
	BatchID    *int64
	TaskID     *int64
	MaterialID *int64
	Type       *int8
}

type inputAllocationRepository struct{}

// NewInputAllocationRepository returns the default MySQL InputAllocationRepository.
func NewInputAllocationRepository(db *sql.DB) InputAllocationRepository {
	return inputAllocationRepository{}
}

func (inputAllocationRepository) Create(ctx context.Context, db domain.DBTX, a *domain.InputAllocation) (int64, error) {
	q := `INSERT INTO input_allocations (batch_id, material_id, task_id, quantity, type, remark, operator_id)
	      VALUES (?, ?, ?, ?, ?, ?, ?)`
	id, err := execInsert(ctx, db, q, a.BatchID, a.MaterialID, a.TaskID, a.Quantity, a.Type, a.Remark, a.OperatorID)
	if err != nil {
		return 0, err
	}
	return id, nil
}

// ReturnableQuantity returns how much of the batch can still be returned for a
// given task: allocated quantity minus already-returned quantity, scoped to the
// same batch AND the same task. Waste (损耗) is unrecoverable, so it never adds
// to the balance — a wasted amount can never be brought back into stock.
func (inputAllocationRepository) ReturnableQuantity(ctx context.Context, db domain.DBTX, batchID int64, taskID *int64) (domain.Decimal, error) {
	// Placeholders appear in order: type (allocate), type (return),
	// batch_id, then task_id only when filtering by a specific task.
	taskCond := "task_id IS NULL"
	args := []any{domain.AllocationTypeAllocate, domain.AllocationTypeReturn, batchID}
	if taskID != nil {
		taskCond = "task_id = ?"
		args = append(args, *taskID)
	}
	q := `SELECT COALESCE(SUM(CASE
	            WHEN type = ? THEN quantity
	            WHEN type = ? THEN -quantity
	            ELSE 0
	          END), 0)
	      FROM input_allocations
	      WHERE batch_id = ? AND ` + taskCond
	var quantity domain.Decimal
	if err := db.QueryRowContext(ctx, q, args...).Scan(&quantity); err != nil {
		return 0, err
	}
	return quantity, nil
}

func (inputAllocationRepository) GetByID(ctx context.Context, db domain.DBTX, id int64) (*domain.InputAllocation, error) {
	q := `SELECT a.id, a.batch_id, b.batch_no, a.material_id, m.name, a.task_id, a.quantity, a.type, a.remark,
	             a.operator_id, u.full_name, a.created_at
	      FROM input_allocations a
	      LEFT JOIN input_batches b ON b.id = a.batch_id
	      LEFT JOIN materials m ON m.id = a.material_id
	      LEFT JOIN users u ON u.id = a.operator_id
	      WHERE a.id=?`
	a := &domain.InputAllocation{}
	err := db.QueryRowContext(ctx, q, id).Scan(&a.ID, &a.BatchID, &a.BatchNo, &a.MaterialID,
		&a.MaterialName, &a.TaskID, &a.Quantity, &a.Type, &a.Remark, &a.OperatorID, &a.OperatorName, &a.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, notFound("领用记录")
	}
	return a, err
}

func (inputAllocationRepository) List(ctx context.Context, db domain.DBTX, p domain.Pagination, f AllocFilter) ([]*domain.InputAllocation, int64, error) {
	var conds []string
	var args []any
	if f.BatchID != nil {
		conds = append(conds, "a.batch_id=?")
		args = append(args, *f.BatchID)
	}
	if f.TaskID != nil {
		conds = append(conds, "a.task_id=?")
		args = append(args, *f.TaskID)
	}
	if f.MaterialID != nil {
		conds = append(conds, "a.material_id=?")
		args = append(args, *f.MaterialID)
	}
	if f.Type != nil {
		conds = append(conds, "a.type=?")
		args = append(args, *f.Type)
	}
	where := ""
	if len(conds) > 0 {
		where = " WHERE " + joinAnd(conds)
	}
	total, err := countRows(ctx, db, `SELECT COUNT(*) FROM input_allocations a`+where, args...)
	if err != nil {
		return nil, 0, err
	}
	q := `SELECT a.id, a.batch_id, b.batch_no, a.material_id, m.name, a.task_id, a.quantity, a.type, a.remark,
	             a.operator_id, u.full_name, a.created_at
	      FROM input_allocations a
	      LEFT JOIN input_batches b ON b.id = a.batch_id
	      LEFT JOIN materials m ON m.id = a.material_id
	      LEFT JOIN users u ON u.id = a.operator_id` + where +
		` ORDER BY a.id DESC LIMIT ? OFFSET ?`
	args = append(args, p.PageSize, p.Offset())
	rows, err := db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out := make([]*domain.InputAllocation, 0)
	for rows.Next() {
		a := &domain.InputAllocation{}
		if err := rows.Scan(&a.ID, &a.BatchID, &a.BatchNo, &a.MaterialID, &a.MaterialName,
			&a.TaskID, &a.Quantity, &a.Type, &a.Remark, &a.OperatorID, &a.OperatorName, &a.CreatedAt); err != nil {
			return nil, 0, err
		}
		out = append(out, a)
	}
	return out, total, rows.Err()
}

func (inputAllocationRepository) ListByTask(ctx context.Context, db domain.DBTX, taskID int64) ([]*domain.InputAllocation, error) {
	q := `SELECT a.id, a.batch_id, b.batch_no, a.material_id, m.name, a.task_id, a.quantity, a.type, a.remark,
	             a.operator_id, u.full_name, a.created_at
	      FROM input_allocations a
	      LEFT JOIN input_batches b ON b.id = a.batch_id
	      LEFT JOIN materials m ON m.id = a.material_id
	      LEFT JOIN users u ON u.id = a.operator_id
	      WHERE a.task_id=? ORDER BY a.id`
	rows, err := db.QueryContext(ctx, q, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]*domain.InputAllocation, 0)
	for rows.Next() {
		a := &domain.InputAllocation{}
		if err := rows.Scan(&a.ID, &a.BatchID, &a.BatchNo, &a.MaterialID, &a.MaterialName,
			&a.TaskID, &a.Quantity, &a.Type, &a.Remark, &a.OperatorID, &a.OperatorName, &a.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

func (inputAllocationRepository) ListByPlan(ctx context.Context, db domain.DBTX, planID int64) ([]*domain.InputAllocation, error) {
	q := `SELECT a.id, a.batch_id, b.batch_no, a.material_id, m.name, a.task_id, a.quantity, a.type, a.remark,
	             a.operator_id, u.full_name, a.created_at
	      FROM input_allocations a
	      LEFT JOIN input_batches b ON b.id = a.batch_id
	      LEFT JOIN materials m ON m.id = a.material_id
	      LEFT JOIN users u ON u.id = a.operator_id
	      LEFT JOIN farm_tasks t ON t.id = a.task_id
	      WHERE t.planting_plan_id=? ORDER BY a.id`
	rows, err := db.QueryContext(ctx, q, planID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]*domain.InputAllocation, 0)
	for rows.Next() {
		a := &domain.InputAllocation{}
		if err := rows.Scan(&a.ID, &a.BatchID, &a.BatchNo, &a.MaterialID, &a.MaterialName,
			&a.TaskID, &a.Quantity, &a.Type, &a.Remark, &a.OperatorID, &a.OperatorName, &a.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}
