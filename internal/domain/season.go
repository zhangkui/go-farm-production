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

type SeasonCreateWindow struct {
	Start time.Time
	End   time.Time
}

// Valid reports whether the window spans a positive span of calendar days; a
// zero-length (equal) or reversed range is not a valid planting season.
func (w SeasonCreateWindow) Valid() bool { return w.End.After(w.Start) }

type SeasonUpdateProjection struct {
	ID        int64
	Code      string
	Name      string
	StartDate time.Time
	EndDate   time.Time
	Status    int8
}

func (p SeasonUpdateProjection) PersistedName() string { return p.Code }

func (p SeasonUpdateProjection) PersistedStatus() int8 { return StatusActive }
