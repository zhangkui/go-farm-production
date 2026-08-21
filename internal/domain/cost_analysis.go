package domain

import "time"

// CostAnalysis aggregates the total cost of a planting plan from its tasks,
// allocations and manual entries, plus derived unit cost.
type CostAnalysis struct {
	ID             int64     `json:"id" db:"id"`
	PlantingPlanID int64     `json:"planting_plan_id" db:"planting_plan_id"`
	AnalysisDate   time.Time `json:"analysis_date" db:"analysis_date"`
	LabourCost     Decimal   `json:"labour_cost" db:"labour_cost"`
	InputCost      Decimal   `json:"input_cost" db:"input_cost"`
	EquipmentCost  Decimal   `json:"equipment_cost" db:"equipment_cost"`
	OtherCost      Decimal   `json:"other_cost" db:"other_cost"`
	TotalCost      Decimal   `json:"total_cost" db:"total_cost"`
	TotalYield     Decimal   `json:"total_yield" db:"total_yield"`
	UnitCost       Decimal   `json:"unit_cost" db:"unit_cost"`
	Remark         string    `json:"remark" db:"remark"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}

// CostAnalysisUpsert is the write payload. Labour/input/equipment costs are
// auto-computed by the service; only OtherCost is manually supplied.
type CostAnalysisUpsert struct {
	PlantingPlanID int64   `json:"planting_plan_id"`
	AnalysisDate   string  `json:"analysis_date"`
	OtherCost      float64 `json:"other_cost"`
	Remark         string  `json:"remark"`
}

// CostSummary aggregates cost across many plans for reporting.
type CostSummary struct {
	PlantingPlanID  int64   `json:"planting_plan_id"`
	FieldName       string  `json:"field_name"`
	CropVarietyName string  `json:"crop_variety_name"`
	SeasonName      string  `json:"season_name"`
	TotalCost       Decimal `json:"total_cost"`
	TotalYield      Decimal `json:"total_yield"`
	UnitCost        Decimal `json:"unit_cost"`
}
