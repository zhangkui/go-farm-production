package domain

import "time"

// CropVariety is the master record of cultivatable species/cultivars.
type CropVariety struct {
	ID          int64     `json:"id" db:"id"`
	Code        string    `json:"code" db:"code"`
	Name        string    `json:"name" db:"name"`
	Category    string    `json:"category" db:"category"`
	GrowthCycle int       `json:"growth_cycle" db:"growth_cycle"`
	Description string    `json:"description" db:"description"`
	Status      int8      `json:"status" db:"status"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// CropVarietyUpsert is the write payload for create/update variety operations.
type CropVarietyUpsert struct {
	Code        string `json:"code"`
	Name        string `json:"name"`
	Category    string `json:"category"`
	GrowthCycle int    `json:"growth_cycle"`
	Description string `json:"description"`
	Status      int8   `json:"status"`
}
