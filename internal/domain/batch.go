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

// RecalculateBatchRemaining derives current stock from immutable ledger
// totals. Allocations and waste both deduct stock irreversibly; returns add
// stock back. The batch quantity therefore must be at least the irreversible
// consumption (allocated - returned + wasted), otherwise the batch has been
// over-consumed and its remaining stock would go negative.
//
// On conflict it returns a 409 AppError and a zero remaining; the caller must
// surface the error and leave the persisted quantity/remaining unchanged.
func RecalculateBatchRemaining(quantity, allocated, returned, wasted Decimal) (Decimal, error) {
	irreversible := allocated.Sub(returned).Add(wasted)
	if quantity < irreversible {
		return 0, Wrap(CodeConflict, 409, "批次数量低于已发生的不可逆消耗（领用-退回+损耗），无法调整", nil)
	}
	return quantity.Sub(irreversible), nil
}
