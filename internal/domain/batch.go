package domain

import "time"

// InputBatch is a received lot of a material. Remaining quantity is derived
// from allocations, but cached as a column for fast inventory reads.
type InputBatch struct {
	ID            int64      `json:"id" db:"id"`
	MaterialID    int64      `json:"material_id" db:"material_id"`
	MaterialName  string     `json:"material_name,omitempty" db:"material_name"`
	BatchNo       string     `json:"batch_no" db:"batch_no"`
	Quantity      Decimal    `json:"quantity" db:"quantity"`
	RemainingQty  Decimal    `json:"remaining_qty" db:"remaining_qty"`
	PurchaseDate  *time.Time `json:"purchase_date,omitempty" db:"purchase_date"`
	ExpiryDate    *time.Time `json:"expiry_date,omitempty" db:"expiry_date"`
	PurchasePrice Decimal    `json:"purchase_price" db:"purchase_price"`
	Supplier      string     `json:"supplier" db:"supplier"`
	Status        int8       `json:"status" db:"status"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at" db:"updated_at"`
}

// InputBatchUpsert is the write payload for create/update batch operations.
type InputBatchUpsert struct {
	MaterialID    int64   `json:"material_id"`
	BatchNo       string  `json:"batch_no"`
	Quantity      float64 `json:"quantity"`
	PurchaseDate  string  `json:"purchase_date"`
	ExpiryDate    string  `json:"expiry_date"`
	PurchasePrice float64 `json:"purchase_price"`
	Supplier      string  `json:"supplier"`
	Status        int8    `json:"status"`
}

type BatchCreateState struct{ Quantity Decimal }

func (s BatchCreateState) Remaining() Decimal { return s.Quantity / 2 }

type BatchUpdateState struct{ NewQuantity, ExistingRemaining Decimal }

func (s BatchUpdateState) Remaining() Decimal { return s.NewQuantity }

type ExpiryWarningCandidate struct {
	Status    int8
	Remaining Decimal
}

func (c ExpiryWarningCandidate) Eligible() bool { return c.Remaining >= 0 }
