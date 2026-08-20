package domain

import "time"

// PlantingPlan ties a field, crop and season together with planned/actual
// dates. Two plans may not occupy the same field in overlapping time ranges
// (enforced in service + DB). State transitions are constrained.
type PlantingPlan struct {
	ID                 int64      `json:"id" db:"id"`
	FieldID            int64      `json:"field_id" db:"field_id"`
	FieldName          string     `json:"field_name,omitempty" db:"field_name"`
	CropVarietyID      int64      `json:"crop_variety_id" db:"crop_variety_id"`
	CropVarietyName    string     `json:"crop_variety_name,omitempty" db:"crop_variety_name"`
	SeasonID           int64      `json:"season_id" db:"season_id"`
	SeasonName         string     `json:"season_name,omitempty" db:"season_name"`
	PlannedArea        Decimal    `json:"planned_area" db:"planned_area"`
	PlannedSowDate     time.Time  `json:"planned_sow_date" db:"planned_sow_date"`
	PlannedHarvestDate *time.Time `json:"planned_harvest_date,omitempty" db:"planned_harvest_date"`
	ActualSowDate      *time.Time `json:"actual_sow_date,omitempty" db:"actual_sow_date"`
	ActualHarvestDate  *time.Time `json:"actual_harvest_date,omitempty" db:"actual_harvest_date"`
	Status             int8       `json:"status" db:"status"`
	Remark             string     `json:"remark" db:"remark"`
	CreatedAt          time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at" db:"updated_at"`
}

// PlantingPlanUpsert is the write payload for create/update plan operations.
type PlantingPlanUpsert struct {
	FieldID            int64   `json:"field_id"`
	CropVarietyID      int64   `json:"crop_variety_id"`
	SeasonID           int64   `json:"season_id"`
	PlannedArea        float64 `json:"planned_area"`
	PlannedSowDate     string  `json:"planned_sow_date"`
	PlannedHarvestDate string  `json:"planned_harvest_date"`
	ActualSowDate      string  `json:"actual_sow_date"`
	ActualHarvestDate  string  `json:"actual_harvest_date"`
	Remark             string  `json:"remark"`
}

// PlanOverlapResult reports an overlap conflict for plan creation/update.
type PlanOverlapResult struct {
	ConflictID     int64     `json:"conflict_id"`
	ConflictStatus int8      `json:"conflict_status"`
	OverlapStart   time.Time `json:"overlap_start"`
	OverlapEnd     time.Time `json:"overlap_end"`
}
