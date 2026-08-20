package domain

import "time"

// InputAllocation records a material movement against a batch and a task:
// allocate (deduct), return (add back), or waste (deduct, unrecoverable).
type InputAllocation struct {
	ID           int64     `json:"id" db:"id"`
	BatchID      int64     `json:"batch_id" db:"batch_id"`
	BatchNo      string    `json:"batch_no,omitempty" db:"batch_no"`
	MaterialID   int64     `json:"material_id" db:"material_id"`
	MaterialName string    `json:"material_name,omitempty" db:"material_name"`
	TaskID       *int64    `json:"task_id,omitempty" db:"task_id"`
	Quantity     Decimal   `json:"quantity" db:"quantity"`
	Type         int8      `json:"type" db:"type"`
	Remark       string    `json:"remark" db:"remark"`
	OperatorID   int64     `json:"operator_id" db:"operator_id"`
	OperatorName string    `json:"operator_name,omitempty" db:"operator_name"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

// InputAllocationUpsert is the write payload for create allocation operations.
type InputAllocationUpsert struct {
	BatchID  int64   `json:"batch_id"`
	TaskID   *int64  `json:"task_id"`
	Quantity float64 `json:"quantity"`
	Type     int8    `json:"type"`
	Remark   string  `json:"remark"`
}
