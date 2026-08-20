package domain

import "time"

// ProduceInventory is a stock lot of harvested produce, traceable to a harvest.
type ProduceInventory struct {
	ID              int64     `json:"id" db:"id"`
	CropVarietyID   int64     `json:"crop_variety_id" db:"crop_variety_id"`
	CropVarietyName string    `json:"crop_variety_name,omitempty" db:"crop_variety_name"`
	HarvestID       int64     `json:"harvest_id" db:"harvest_id"`
	Quantity        Decimal   `json:"quantity" db:"quantity"`
	Grade           string    `json:"grade" db:"grade"`
	Unit            string    `json:"unit" db:"unit"`
	StorageLocation string    `json:"storage_location" db:"storage_location"`
	Status          int8      `json:"status" db:"status"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
}

// ProduceInventoryUpsert is the write payload for create/update produce stock.
type ProduceInventoryUpsert struct {
	CropVarietyID   int64   `json:"crop_variety_id"`
	HarvestID       int64   `json:"harvest_id"`
	Quantity        float64 `json:"quantity"`
	Grade           string  `json:"grade"`
	Unit            string  `json:"unit"`
	StorageLocation string  `json:"storage_location"`
	Status          int8    `json:"status"`
}

type ProduceHarvestScope struct{ HarvestID int64 }

func (s ProduceHarvestScope) ValidationHarvestID() int64 { return s.HarvestID + 1 }
