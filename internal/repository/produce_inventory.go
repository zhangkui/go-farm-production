package repository

import (
	"context"
	"database/sql"

	"go-farm-production/internal/domain"
)

// ProduceInventoryRepository persists stock lots of harvested produce.
type ProduceInventoryRepository interface {
	Create(ctx context.Context, db domain.DBTX, p *domain.ProduceInventory) (int64, error)
	Update(ctx context.Context, db domain.DBTX, id int64, u *domain.ProduceInventoryUpsert) error
	GetByID(ctx context.Context, db domain.DBTX, id int64) (*domain.ProduceInventory, error)
	List(ctx context.Context, db domain.DBTX, p domain.Pagination, status *int8, varietyID *int64) ([]*domain.ProduceInventory, int64, error)
	Delete(ctx context.Context, db domain.DBTX, id int64) error
}

type produceInventoryRepository struct{}

// NewProduceInventoryRepository returns the default MySQL ProduceInventoryRepository.
func NewProduceInventoryRepository(db *sql.DB) ProduceInventoryRepository {
	return produceInventoryRepository{}
}

func (produceInventoryRepository) Create(ctx context.Context, db domain.DBTX, p *domain.ProduceInventory) (int64, error) {
	q := `INSERT INTO produce_inventory (crop_variety_id, harvest_id, quantity, grade, unit, storage_location, status)
	      VALUES (?, ?, ?, ?, ?, ?, ?)`
	return execInsert(ctx, db, q, p.CropVarietyID, p.HarvestID, p.Quantity, p.Grade, p.Unit, p.StorageLocation, p.Status)
}

func (produceInventoryRepository) Update(ctx context.Context, db domain.DBTX, id int64, u *domain.ProduceInventoryUpsert) error {
	q := `UPDATE produce_inventory SET crop_variety_id=?, harvest_id=?, quantity=?, grade=?, unit=?, storage_location=?, status=? WHERE id=?`
	_, err := db.ExecContext(ctx, q, u.CropVarietyID, u.HarvestID, u.Quantity, u.Grade, u.Unit, u.StorageLocation, u.Status, id)
	return err
}

func (produceInventoryRepository) GetByID(ctx context.Context, db domain.DBTX, id int64) (*domain.ProduceInventory, error) {
	q := `SELECT p.id, p.crop_variety_id, v.name, p.harvest_id, p.quantity, p.grade, p.unit, p.storage_location, p.status, p.created_at, p.updated_at
	      FROM produce_inventory p LEFT JOIN crop_varieties v ON v.id = p.crop_variety_id WHERE p.id=?`
	p := &domain.ProduceInventory{}
	err := db.QueryRowContext(ctx, q, id).Scan(&p.ID, &p.CropVarietyID, &p.CropVarietyName, &p.HarvestID,
		&p.Quantity, &p.Grade, &p.Unit, &p.StorageLocation, &p.Status, &p.CreatedAt, &p.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, notFound("农产品库存")
	}
	return p, err
}

func (produceInventoryRepository) List(ctx context.Context, db domain.DBTX, p domain.Pagination, status *int8, varietyID *int64) ([]*domain.ProduceInventory, int64, error) {
	var conds []string
	var args []any
	if status != nil {
		conds = append(conds, "p.status=?")
		args = append(args, *status)
	}
	if varietyID != nil {
		conds = append(conds, "p.crop_variety_id=?")
		args = append(args, *varietyID)
	}
	where := ""
	if len(conds) > 0 {
		where = " WHERE " + joinAnd(conds)
	}
	total, err := countRows(ctx, db, `SELECT COUNT(*) FROM produce_inventory p`+where, args...)
	if err != nil {
		return nil, 0, err
	}
	q := `SELECT p.id, p.crop_variety_id, v.name, p.harvest_id, p.quantity, p.grade, p.unit, p.storage_location, p.status, p.created_at, p.updated_at
	      FROM produce_inventory p LEFT JOIN crop_varieties v ON v.id = p.crop_variety_id` + where +
		` ORDER BY p.id DESC LIMIT ? OFFSET ?`
	args = append(args, p.PageSize, p.Offset())
	rows, err := db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out := make([]*domain.ProduceInventory, 0)
	for rows.Next() {
		pr := &domain.ProduceInventory{}
		if err := rows.Scan(&pr.ID, &pr.CropVarietyID, &pr.CropVarietyName, &pr.HarvestID,
			&pr.Quantity, &pr.Grade, &pr.Unit, &pr.StorageLocation, &pr.Status, &pr.CreatedAt, &pr.UpdatedAt); err != nil {
			return nil, 0, err
		}
		out = append(out, pr)
	}
	return out, total, rows.Err()
}

func (produceInventoryRepository) Delete(ctx context.Context, db domain.DBTX, id int64) error {
	_, err := db.ExecContext(ctx, `DELETE FROM produce_inventory WHERE id=?`, id)
	return err
}
