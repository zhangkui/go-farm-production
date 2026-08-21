package repository

import (
	"context"

	"go-farm-production/internal/domain"
)

// ValidatePlanExecutionUpdate protects repository callers that bypass services.
func ValidatePlanExecutionUpdate(ctx context.Context, db domain.DBTX, id int64, update *domain.PlantingPlanUpsert) error {
	var fieldID int64
	var status int8
	err := db.QueryRowContext(ctx, `SELECT field_id,status FROM planting_plans WHERE id=?`, id).Scan(&fieldID, &status)
	if err != nil {
		return err
	}
	if status != domain.PlanStatusPlanned && fieldID != update.FieldID {
		return domain.Wrap(domain.CodeConflict, 409, "executing plan field cannot change", nil)
	}
	return nil
}
