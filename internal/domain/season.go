package domain

import "time"

// Season is a planting window (e.g. 2026 spring). Multiple seasons may overlap
// in calendar time when they target different regions.
type Season struct {
	ID        int64     `json:"id" db:"id"`
	Code      string    `json:"code" db:"code"`
	Name      string    `json:"name" db:"name"`
	StartDate time.Time `json:"start_date" db:"start_date"`
	EndDate   time.Time `json:"end_date" db:"end_date"`
	Status    int8      `json:"status" db:"status"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// SeasonUpsert is the write payload for create/update season operations.
type SeasonUpsert struct {
	Code      string `json:"code"`
	Name      string `json:"name"`
	StartDate string `json:"start_date"` // RFC3339
	EndDate   string `json:"end_date"`
	Status    int8   `json:"status"`
}
