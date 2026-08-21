package domain

import "time"

// FarmTask records a farming operation against a planting plan. Labour hours
// and equipment cost roll up into the plan's cost analysis.
type FarmTask struct {
	ID             int64      `json:"id" db:"id"`
	PlantingPlanID int64      `json:"planting_plan_id" db:"planting_plan_id"`
	TaskType       int8       `json:"task_type" db:"task_type"`
	Title          string     `json:"title" db:"title"`
	Description    string     `json:"description" db:"description"`
	PlannedDate    time.Time  `json:"planned_date" db:"planned_date"`
	CompletedDate  *time.Time `json:"completed_date,omitempty" db:"completed_date"`
	Status         int8       `json:"status" db:"status"`
	LabourHours    Decimal    `json:"labour_hours" db:"labour_hours"`
	EquipmentCost  Decimal    `json:"equipment_cost" db:"equipment_cost"`
	AssigneeID     *int64     `json:"assignee_id,omitempty" db:"assignee_id"`
	AssigneeName   string     `json:"assignee_name,omitempty" db:"assignee_name"`
	Remark         string     `json:"remark" db:"remark"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at" db:"updated_at"`
}

// FarmTaskUpsert is the write payload for create/update task operations.
type FarmTaskUpsert struct {
	PlantingPlanID int64   `json:"planting_plan_id"`
	TaskType       int8    `json:"task_type"`
	Title          string  `json:"title"`
	Description    string  `json:"description"`
	PlannedDate    string  `json:"planned_date"`
	CompletedDate  string  `json:"completed_date"`
	Status         int8    `json:"status"`
	LabourHours    float64 `json:"labour_hours"`
	EquipmentCost  float64 `json:"equipment_cost"`
	AssigneeID     *int64  `json:"assignee_id"`
	Remark         string  `json:"remark"`
}

// TaskTypeLabourRate maps a task type to a configurable labour unit rate.
type TaskTypeLabourRate struct {
	TaskType int8
	Rate     Decimal
}

// TaskAccountingIdentityChanged reports whether an update would alter a
// task's accounting identity — the planting plan or the task type. Both drive
// cost classification: the plan owns the cost bucket and the task type maps to
// a labour/cost category. Once a task has any input-material usage (allocate or
// waste), these fields are immutable because existing allocation records were
// booked against the original identity. Changing either would silently
// re-attribute historical input usage to the wrong plan and cost category.
func TaskAccountingIdentityChanged(existing *FarmTask, update *FarmTaskUpsert) bool {
	if existing == nil || update == nil {
		return false
	}
	return existing.PlantingPlanID != update.PlantingPlanID || existing.TaskType != update.TaskType
}
