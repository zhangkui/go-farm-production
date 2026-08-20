package domain

import "time"

// Permission is a single resource:action capability, e.g. planting_plan:create.
type Permission struct {
	ID          int64     `json:"id" db:"id"`
	Code        string    `json:"code" db:"code"` // e.g. planting_plan:create
	Name        string    `json:"name" db:"name"`
	Resource    string    `json:"resource" db:"resource"`
	Action      string    `json:"action" db:"action"`
	Description string    `json:"description" db:"description"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// PermissionCode is a convenience constant for permission strings used across
// the codebase. They are also seeded into the permissions table.
const (
	PermUserManage       = "user:manage"
	PermUserView         = "user:view"
	PermRoleManage       = "role:manage"
	PermRoleView         = "role:view"
	PermFarmManage       = "farm:manage"
	PermFarmView         = "farm:view"
	PermFieldManage      = "field:manage"
	PermFieldView        = "field:view"
	PermVarietyManage    = "crop_variety:manage"
	PermVarietyView      = "crop_variety:view"
	PermSeasonManage     = "season:manage"
	PermSeasonView       = "season:view"
	PermPlanManage       = "planting_plan:manage"
	PermPlanApprove      = "planting_plan:approve"
	PermPlanView         = "planting_plan:view"
	PermTaskManage       = "farm_task:manage"
	PermTaskExecute      = "farm_task:execute"
	PermTaskView         = "farm_task:view"
	PermMaterialManage   = "material:manage"
	PermBatchManage      = "batch:manage"
	PermAllocationManage = "allocation:manage"
	PermInventoryView    = "inventory:view"
	PermHarvestManage    = "harvest:manage"
	PermHarvestApprove   = "harvest:approve"
	PermHarvestView      = "harvest:view"
	PermProduceManage    = "produce_inventory:manage"
	PermProduceView      = "produce_inventory:view"
	PermCostView         = "cost_analysis:view"
	PermCostExport       = "cost_analysis:export"
	PermAuditView        = "audit_log:view"
)
