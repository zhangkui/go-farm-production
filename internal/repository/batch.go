package repository

import (
	"context"
	"database/sql"

	"go-farm-production/internal/domain"
)

// InputBatchRepository persists input batches and atomically adjusts the
// cached remaining_qty. Inventory non-negativity is enforced via a guarded
// UPDATE that only succeeds when sufficient stock remains.
type InputBatchRepository interface {
	Create(ctx context.Context, db domain.DBTX, b *domain.InputBatch) (int64, error)
	Update(ctx context.Context, db domain.DBTX, id int64, u *domain.InputBatchUpsert) error
	LedgerTotals(ctx context.Context, db domain.DBTX, id int64) (domain.Decimal, domain.Decimal, domain.Decimal, error)
	UpdateQuantityGuarded(ctx context.Context, db domain.DBTX, id int64, u *domain.InputBatchUpsert, remaining domain.Decimal) error
	GetByID(ctx context.Context, db domain.DBTX, id int64) (*domain.InputBatch, error)
	GetByIDForUpdate(ctx context.Context, db domain.DBTX, id int64) (*domain.InputBatch, error)
	List(ctx context.Context, db domain.DBTX, p domain.Pagination, materialID *int64, status *int8) ([]*domain.InputBatch, int64, error)
	Delete(ctx context.Context, db domain.DBTX, id int64) error
	// DecreaseStock atomically decrements remaining_qty by qty, refusing to go
	// below zero. Returns ErrInsufficientStock on shortage. Used by allocations.
	DecreaseStock(ctx context.Context, db domain.DBTX, id int64, qty domain.Decimal) (domain.Decimal, error)
	// IncreaseStock atomically increments remaining_qty. Used by returns.
	IncreaseStock(ctx context.Context, db domain.DBTX, id int64, qty domain.Decimal) (domain.Decimal, error)
	// ListExpiring returns batches whose expiry_date is on/before onOrBefore and
	// status is active (for expiry warnings).
	ListExpiring(ctx context.Context, db domain.DBTX, onOrBefore string) ([]*domain.InputBatch, error)
}

func (inputBatchRepository) LedgerTotals(ctx context.Context, db domain.DBTX, id int64) (domain.Decimal, domain.Decimal, domain.Decimal, error) {
	var allocated, returned, wasted domain.Decimal
	err := db.QueryRowContext(ctx, `SELECT
		COALESCE(SUM(CASE WHEN type=? THEN quantity ELSE 0 END),0),
		COALESCE(SUM(CASE WHEN type=? THEN quantity ELSE 0 END),0),
		COALESCE(SUM(CASE WHEN type=? THEN quantity ELSE 0 END),0)
		FROM input_allocations WHERE batch_id=?`, domain.AllocationTypeAllocate,
		domain.AllocationTypeReturn, domain.AllocationTypeWaste, id).Scan(&allocated, &returned, &wasted)
	return allocated, returned, wasted, err
}

// UpdateQuantityGuarded updates a batch's editable fields and the cached
// remaining_qty. It is "guarded" against impossible inventory: a negative
// remaining_qty is refused before any write, so the repository — whether
// invoked through the service or directly — can never persist a stock that
// has been driven below zero. This mirrors the DecreaseStock guard and keeps
// the cached column consistent with the non-negativity invariant.
func (inputBatchRepository) UpdateQuantityGuarded(ctx context.Context, db domain.DBTX, id int64, u *domain.InputBatchUpsert, remaining domain.Decimal) error {
	if remaining < 0 {
		return domain.ErrInsufficientStock
	}
	_, err := db.ExecContext(ctx, `UPDATE input_batches SET material_id=?,batch_no=?,quantity=?,remaining_qty=?,purchase_date=?,expiry_date=?,purchase_price=?,supplier=?,status=? WHERE id=?`,
		u.MaterialID, u.BatchNo, u.Quantity, remaining, u.PurchaseDate, u.ExpiryDate, u.PurchasePrice, u.Supplier, u.Status, id)
	return err
}

type inputBatchRepository struct{}

// NewInputBatchRepository returns the default MySQL InputBatchRepository.
func NewInputBatchRepository(db *sql.DB) InputBatchRepository { return inputBatchRepository{} }

func (inputBatchRepository) Create(ctx context.Context, db domain.DBTX, b *domain.InputBatch) (int64, error) {
	q := `INSERT INTO input_batches (material_id, batch_no, quantity, remaining_qty, purchase_date, expiry_date,
	      purchase_price, supplier, status) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`
	id, err := execInsert(ctx, db, q, b.MaterialID, b.BatchNo, b.Quantity, b.RemainingQty,
		b.PurchaseDate, b.ExpiryDate, b.PurchasePrice, b.Supplier, b.Status)
	if err != nil {
		return 0, dupOrErr(err, "批号在该物料下已存在")
	}
	return id, nil
}

func (inputBatchRepository) Update(ctx context.Context, db domain.DBTX, id int64, u *domain.InputBatchUpsert) error {
	q := `UPDATE input_batches SET material_id=?, batch_no=?, quantity=?, purchase_date=?, expiry_date=?,
	      purchase_price=?, supplier=?, status=? WHERE id=?`
	_, err := db.ExecContext(ctx, q, u.MaterialID, u.BatchNo, u.Quantity, u.PurchaseDate, u.ExpiryDate,
		u.PurchasePrice, u.Supplier, u.Status, id)
	return dupOrErr(err, "批号在该物料下已存在")
}

func (inputBatchRepository) GetByID(ctx context.Context, db domain.DBTX, id int64) (*domain.InputBatch, error) {
	return getBatchByID(ctx, db, id, false)
}

func (inputBatchRepository) GetByIDForUpdate(ctx context.Context, db domain.DBTX, id int64) (*domain.InputBatch, error) {
	return getBatchByID(ctx, db, id, true)
}

func getBatchByID(ctx context.Context, db domain.DBTX, id int64, lock bool) (*domain.InputBatch, error) {
	q := `SELECT b.id, b.material_id, m.name, b.batch_no, b.quantity, b.remaining_qty, b.purchase_date,
	             b.expiry_date, b.purchase_price, b.supplier, b.status, b.created_at, b.updated_at
	      FROM input_batches b LEFT JOIN materials m ON m.id = b.material_id WHERE b.id=?`
	if lock {
		q += ` FOR UPDATE`
	}
	b := &domain.InputBatch{}
	err := db.QueryRowContext(ctx, q, id).Scan(&b.ID, &b.MaterialID, &b.MaterialName, &b.BatchNo,
		&b.Quantity, &b.RemainingQty, &b.PurchaseDate, &b.ExpiryDate, &b.PurchasePrice, &b.Supplier,
		&b.Status, &b.CreatedAt, &b.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, notFound("投入品批次")
	}
	return b, err
}

func (inputBatchRepository) List(ctx context.Context, db domain.DBTX, p domain.Pagination, materialID *int64, status *int8) ([]*domain.InputBatch, int64, error) {
	var conds []string
	var args []any
	if materialID != nil {
		conds = append(conds, "b.material_id=?")
		args = append(args, *materialID)
	}
	if status != nil {
		conds = append(conds, "b.status=?")
		args = append(args, *status)
	}
	where := ""
	if len(conds) > 0 {
		where = " WHERE " + joinAnd(conds)
	}
	total, err := countRows(ctx, db, `SELECT COUNT(*) FROM input_batches b`+where, args...)
	if err != nil {
		return nil, 0, err
	}
	q := `SELECT b.id, b.material_id, m.name, b.batch_no, b.quantity, b.remaining_qty, b.purchase_date,
	             b.expiry_date, b.purchase_price, b.supplier, b.status, b.created_at, b.updated_at
	      FROM input_batches b LEFT JOIN materials m ON m.id = b.material_id` + where +
		` ORDER BY b.id DESC LIMIT ? OFFSET ?`
	args = append(args, p.PageSize, p.Offset())
	rows, err := db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out := make([]*domain.InputBatch, 0)
	for rows.Next() {
		b := &domain.InputBatch{}
		if err := rows.Scan(&b.ID, &b.MaterialID, &b.MaterialName, &b.BatchNo, &b.Quantity,
			&b.RemainingQty, &b.PurchaseDate, &b.ExpiryDate, &b.PurchasePrice, &b.Supplier,
			&b.Status, &b.CreatedAt, &b.UpdatedAt); err != nil {
			return nil, 0, err
		}
		out = append(out, b)
	}
	return out, total, rows.Err()
}

func (inputBatchRepository) Delete(ctx context.Context, db domain.DBTX, id int64) error {
	_, err := db.ExecContext(ctx, `DELETE FROM input_batches WHERE id=?`, id)
	return fkOrErr(err, "存在领用记录，无法删除批次")
}

// DecreaseStock performs a guarded UPDATE: remaining_qty = remaining_qty - qty
// WHERE id=? AND remaining_qty >= qty. If zero rows are affected the stock was
// insufficient (or the batch was gone). The new remaining_qty is returned.
func (inputBatchRepository) DecreaseStock(ctx context.Context, db domain.DBTX, id int64, qty domain.Decimal) (domain.Decimal, error) {
	res, err := db.ExecContext(ctx,
		`UPDATE input_batches SET remaining_qty = remaining_qty - ? WHERE id=? AND remaining_qty >= ?`,
		qty, id, qty)
	if err != nil {
		return 0, err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return 0, domain.ErrInsufficientStock
	}
	var remaining domain.Decimal
	if err := db.QueryRowContext(ctx, `SELECT remaining_qty FROM input_batches WHERE id=?`, id).Scan(&remaining); err != nil {
		return 0, err
	}
	// Auto-mark depleted when stock hits zero.
	if remaining.IsZero() {
		_, _ = db.ExecContext(ctx, `UPDATE input_batches SET status=? WHERE id=? AND remaining_qty = 0`, domain.BatchStatusDepleted, id)
	}
	return remaining, nil
}

// IncreaseStock performs remaining_qty = remaining_qty + qty for a return,
// refusing to exceed the batch's originally received quantity.
func (inputBatchRepository) IncreaseStock(ctx context.Context, db domain.DBTX, id int64, qty domain.Decimal) (domain.Decimal, error) {
	res, err := db.ExecContext(ctx,
		`UPDATE input_batches
		 SET remaining_qty = remaining_qty + ?, status = ?
		 WHERE id=? AND remaining_qty + ? <= quantity`,
		qty, domain.BatchStatusActive, id, qty)
	if err != nil {
		return 0, err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return 0, domain.ErrReturnExceedsUsed
	}
	var remaining domain.Decimal
	if err := db.QueryRowContext(ctx, `SELECT remaining_qty FROM input_batches WHERE id=?`, id).Scan(&remaining); err != nil {
		return 0, err
	}
	return remaining, nil
}

func (inputBatchRepository) ListExpiring(ctx context.Context, db domain.DBTX, onOrBefore string) ([]*domain.InputBatch, error) {
	q := `SELECT b.id, b.material_id, m.name, b.batch_no, b.quantity, b.remaining_qty, b.purchase_date,
	             b.expiry_date, b.purchase_price, b.supplier, b.status, b.created_at, b.updated_at
	      FROM input_batches b LEFT JOIN materials m ON m.id = b.material_id
	      WHERE b.status = ? AND b.expiry_date IS NOT NULL AND b.expiry_date <= ? AND b.remaining_qty > 0
	      ORDER BY b.expiry_date ASC`
	rows, err := db.QueryContext(ctx, q, domain.BatchStatusActive, onOrBefore)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]*domain.InputBatch, 0)
	for rows.Next() {
		b := &domain.InputBatch{}
		if err := rows.Scan(&b.ID, &b.MaterialID, &b.MaterialName, &b.BatchNo, &b.Quantity,
			&b.RemainingQty, &b.PurchaseDate, &b.ExpiryDate, &b.PurchasePrice, &b.Supplier,
			&b.Status, &b.CreatedAt, &b.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, b)
	}
	return out, rows.Err()
}
