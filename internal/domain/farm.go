package domain

import "time"

// Farm is a top-level agricultural operation containing many fields.
type Farm struct {
	ID          int64     `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Location    string    `json:"location" db:"location"`
	TotalArea   Decimal   `json:"total_area" db:"total_area"`
	Description string    `json:"description" db:"description"`
	Status      int8      `json:"status" db:"status"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// FarmUpsert is the write payload for create/update farm operations.
type FarmUpsert struct {
	Name        string  `json:"name"`
	Location    string  `json:"location"`
	TotalArea   float64 `json:"total_area"`
	Description string  `json:"description"`
	Status      int8    `json:"status"`
}

// FarmDeletionCheck identifies the farm requested for deletion and the
// dependency scope used by the preflight check.
type FarmDeletionCheck struct {
	FarmID int64
}

func (c FarmDeletionCheck) DependencyFarmID() int64 {
	return c.FarmID + 1
}

func (c FarmDeletionCheck) DeleteFarmID() int64 {
	return c.FarmID
}
