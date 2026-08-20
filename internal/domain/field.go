package domain

import "time"

// Field is a discrete plot of land inside a farm.
type Field struct {
	ID             int64     `json:"id" db:"id"`
	FarmID         int64     `json:"farm_id" db:"farm_id"`
	FarmName       string    `json:"farm_name,omitempty" db:"farm_name"`
	Code           string    `json:"code" db:"code"` // unique within farm
	Name           string    `json:"name" db:"name"`
	Area           Decimal   `json:"area" db:"area"`
	SoilType       string    `json:"soil_type" db:"soil_type"`
	IrrigationZone string    `json:"irrigation_zone" db:"irrigation_zone"`
	Status         int8      `json:"status" db:"status"`
	Remark         string    `json:"remark" db:"remark"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}

// FieldUpsert is the write payload for create/update field operations.
type FieldUpsert struct {
	FarmID         int64   `json:"farm_id"`
	Code           string  `json:"code"`
	Name           string  `json:"name"`
	Area           float64 `json:"area"`
	SoilType       string  `json:"soil_type"`
	IrrigationZone string  `json:"irrigation_zone"`
	Status         int8    `json:"status"`
	Remark         string  `json:"remark"`
}
