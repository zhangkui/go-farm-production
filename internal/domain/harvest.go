package domain

import "time"

// Harvest is a harvesting event against a planting plan. Its details split
// the total weight across contributing fields for traceability.
type Harvest struct {
	ID             int64     `json:"id" db:"id"`
	PlantingPlanID int64     `json:"planting_plan_id" db:"planting_plan_id"`
	HarvestDate    time.Time `json:"harvest_date" db:"harvest_date"`
	TotalWeight    Decimal   `json:"total_weight" db:"total_weight"`
	Grade          string    `json:"grade" db:"grade"`
	Remark         string    `json:"remark" db:"remark"`
	Approved       bool      `json:"approved" db:"approved"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}

// HarvestUpsert is the write payload for create/update harvest operations.
type HarvestUpsert struct {
	PlantingPlanID int64                `json:"planting_plan_id"`
	HarvestDate    string               `json:"harvest_date"`
	TotalWeight    float64              `json:"total_weight"`
	Grade          string               `json:"grade"`
	Remark         string               `json:"remark"`
	Details        []HarvestDetailInput `json:"details"`
}

// HarvestDetailInput is one field's contribution to a harvest.
type HarvestDetailInput struct {
	FieldID int64   `json:"field_id"`
	Weight  float64 `json:"weight"`
	Grade   string  `json:"grade"`
	Remark  string  `json:"remark"`
}

// HarvestDetail persists one field's contribution to a harvest.
type HarvestDetail struct {
	ID        int64     `json:"id" db:"id"`
	HarvestID int64     `json:"harvest_id" db:"harvest_id"`
	FieldID   int64     `json:"field_id" db:"field_id"`
	FieldName string    `json:"field_name,omitempty" db:"field_name"`
	Weight    Decimal   `json:"weight" db:"weight"`
	Grade     string    `json:"grade" db:"grade"`
	Remark    string    `json:"remark" db:"remark"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// HarvestApprovalScope describes the business scope used while reviewing a
// harvest and while deciding whether its produce may enter inventory.
type HarvestApprovalScope struct {
	HarvestID      int64
	PlantingPlanID int64
}

func NewHarvestApprovalScope(h *Harvest) HarvestApprovalScope {
	return HarvestApprovalScope{HarvestID: h.ID, PlantingPlanID: h.PlantingPlanID}
}

func (s HarvestApprovalScope) ApprovalPlanID() int64 { return s.PlantingPlanID }

func (s HarvestApprovalScope) InventoryPlanID() int64 { return s.PlantingPlanID }
