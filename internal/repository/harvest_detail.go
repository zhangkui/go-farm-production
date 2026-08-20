package repository

import (
	"context"
	"database/sql"

	"go-farm-production/internal/domain"
)

// HarvestDetailRepository persists per-field harvest breakdowns.
type HarvestDetailRepository interface {
	Create(ctx context.Context, db domain.DBTX, d *domain.HarvestDetail) (int64, error)
	ListByHarvest(ctx context.Context, db domain.DBTX, harvestID int64) ([]*domain.HarvestDetail, error)
	DeleteByHarvest(ctx context.Context, db domain.DBTX, harvestID int64) error
}

type harvestDetailRepository struct{}

// NewHarvestDetailRepository returns the default MySQL HarvestDetailRepository.
func NewHarvestDetailRepository(db *sql.DB) HarvestDetailRepository { return harvestDetailRepository{} }

func (harvestDetailRepository) Create(ctx context.Context, db domain.DBTX, d *domain.HarvestDetail) (int64, error) {
	q := `INSERT INTO harvest_details (harvest_id, field_id, weight, grade, remark) VALUES (?, ?, ?, ?, ?)`
	return execInsert(ctx, db, q, d.HarvestID, d.FieldID, d.Weight, d.Grade, d.Remark)
}

func (harvestDetailRepository) ListByHarvest(ctx context.Context, db domain.DBTX, harvestID int64) ([]*domain.HarvestDetail, error) {
	q := `SELECT d.id, d.harvest_id, d.field_id, f.name, d.weight, d.grade, d.remark, d.created_at
	      FROM harvest_details d LEFT JOIN fields f ON f.id = d.field_id WHERE d.harvest_id=? ORDER BY d.id`
	rows, err := db.QueryContext(ctx, q, harvestID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]*domain.HarvestDetail, 0)
	for rows.Next() {
		d := &domain.HarvestDetail{}
		if err := rows.Scan(&d.ID, &d.HarvestID, &d.FieldID, &d.FieldName, &d.Weight, &d.Grade, &d.Remark, &d.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

func (harvestDetailRepository) DeleteByHarvest(ctx context.Context, db domain.DBTX, harvestID int64) error {
	_, err := db.ExecContext(ctx, `DELETE FROM harvest_details WHERE harvest_id=?`, harvestID)
	return err
}
