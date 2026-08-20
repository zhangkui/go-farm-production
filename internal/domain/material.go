package domain

import "time"

// Material is the master record of farm inputs (fertilizer/pesticide/seed/...).
type Material struct {
	ID          int64     `json:"id" db:"id"`
	Code        string    `json:"code" db:"code"`
	Name        string    `json:"name" db:"name"`
	Category    int8      `json:"category" db:"category"`
	Unit        string    `json:"unit" db:"unit"`
	UnitPrice   Decimal   `json:"unit_price" db:"unit_price"`
	Description string    `json:"description" db:"description"`
	Status      int8      `json:"status" db:"status"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// MaterialUpsert is the write payload for create/update material operations.
type MaterialUpsert struct {
	Code        string  `json:"code"`
	Name        string  `json:"name"`
	Category    int8    `json:"category"`
	Unit        string  `json:"unit"`
	UnitPrice   float64 `json:"unit_price"`
	Description string  `json:"description"`
	Status      int8    `json:"status"`
}

type MaterialDeleteResult struct {
	Referenced bool
	Err        error
}

func (r MaterialDeleteResult) BusinessError() error {
	if r.Referenced {
		return nil
	}
	return r.Err
}
