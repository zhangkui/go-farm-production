package verify

import (
	"context"
	"testing"

	"go-farm-production/internal/domain"
)

func TestBug031_BusinessRegression(t *testing.T) {
	env := newVerifyEnv(t, "BUG-031")
	ctx := context.Background()
	varietyID, planID, otherPlanID := bug031PlanningFixture(t, env)
	harvestA := bug031Harvest(t, env, planID, "2026-08-01", 10)
	harvestB := bug031Harvest(t, env, planID, "2026-08-02", 12)
	harvestC := bug031Harvest(t, env, otherPlanID, "2026-08-03", 14)

	if err := env.Services.Harvest.Approve(ctx, harvestA); err != nil {
		t.Fatalf("approve target harvest: %v", err)
	}
	assertBug031Approval(t, env, harvestA, 1, "target harvest")
	assertBug031Approval(t, env, harvestB, 0, "same-plan sibling harvest")
	assertBug031Approval(t, env, harvestC, 0, "other-plan harvest")

	// Reset the sibling explicitly so the inventory assertion isolates the
	// second boundary: approval must belong to the requested harvest itself.
	execVerifySQL(t, env.DB, `UPDATE harvests SET approved=0 WHERE id=?`, harvestB)
	illegalID, err := env.Services.Produce.Create(ctx, &domain.ProduceInventoryUpsert{
		CropVarietyID:   varietyID,
		HarvestID:       harvestB,
		Quantity:        5,
		Grade:           "A",
		Unit:            "kg",
		StorageLocation: "cold-room-b",
		Status:          domain.ProduceStatusInStock,
	})
	if err == nil {
		t.Errorf("unapproved sibling harvest entered inventory with id=%d", illegalID)
	} else if appErr := domain.AsAppError(err); appErr.Code != domain.CodeConflict {
		t.Errorf("unapproved sibling error code=%d, want %d: %v", appErr.Code, domain.CodeConflict, err)
	}
	if got := queryVerifyInt(t, env.DB, `SELECT COUNT(*) FROM produce_inventory WHERE harvest_id=?`, harvestB); got != 0 {
		t.Errorf("unapproved sibling inventory rows=%d, want 0", got)
	}
	if got := queryVerifyInt(t, env.DB,
		`SELECT COUNT(*) FROM audit_logs WHERE action='create' AND resource_type='produce_inventory'`); got != 0 {
		t.Errorf("failed inventory request wrote %d success audits, want 0", got)
	}

	validID, err := env.Services.Produce.Create(ctx, &domain.ProduceInventoryUpsert{
		CropVarietyID:   varietyID,
		HarvestID:       harvestA,
		Quantity:        6,
		Grade:           "A",
		Unit:            "kg",
		StorageLocation: "cold-room-a",
		Status:          domain.ProduceStatusInStock,
	})
	if err != nil {
		t.Fatalf("approved target harvest inventory failed: %v", err)
	}
	if validID == 0 {
		t.Errorf("approved target harvest returned zero inventory id")
	}

	if _, err := env.Services.Produce.Create(ctx, &domain.ProduceInventoryUpsert{
		CropVarietyID:   varietyID,
		HarvestID:       harvestC,
		Quantity:        4,
		Grade:           "B",
		Unit:            "kg",
		StorageLocation: "cold-room-c",
		Status:          domain.ProduceStatusInStock,
	}); err == nil {
		t.Errorf("unapproved other-plan harvest entered inventory")
	} else if appErr := domain.AsAppError(err); appErr.Code != domain.CodeConflict {
		t.Errorf("unapproved other-plan error code=%d, want %d: %v", appErr.Code, domain.CodeConflict, err)
	}
	if got := queryVerifyInt(t, env.DB, `SELECT COUNT(*) FROM produce_inventory WHERE harvest_id=?`, harvestC); got != 0 {
		t.Errorf("unapproved other-plan inventory rows=%d, want 0", got)
	}
	assertBug031Approval(t, env, harvestB, 0, "same-plan sibling after inventory checks")
	assertBug031Approval(t, env, harvestC, 0, "other-plan harvest after inventory checks")
}

func bug031PlanningFixture(t *testing.T, env *verifyEnv) (int64, int64, int64) {
	t.Helper()
	farmID := lastVerifyID(t, execVerifySQL(t, env.DB,
		`INSERT INTO farms (name, location, total_area, status) VALUES ('BUG031 farm', 'north', 100, 1)`))
	fieldID := lastVerifyID(t, execVerifySQL(t, env.DB,
		`INSERT INTO fields (farm_id, code, name, area, status) VALUES (?, 'BUG031-F1', 'field one', 40, 1)`, farmID))
	otherFieldID := lastVerifyID(t, execVerifySQL(t, env.DB,
		`INSERT INTO fields (farm_id, code, name, area, status) VALUES (?, 'BUG031-F2', 'field two', 40, 1)`, farmID))
	varietyID := lastVerifyID(t, execVerifySQL(t, env.DB,
		`INSERT INTO crop_varieties (code, name, category, growth_cycle, description, status) VALUES ('BUG031-V', 'rice', 'grain', 100, '', 1)`))
	seasonID := lastVerifyID(t, execVerifySQL(t, env.DB,
		`INSERT INTO seasons (code, name, start_date, end_date, status) VALUES ('BUG031-S', 'summer', '2026-05-01', '2026-10-01', 2)`))
	planID := lastVerifyID(t, execVerifySQL(t, env.DB,
		`INSERT INTO planting_plans (field_id, crop_variety_id, season_id, planned_area, planned_sow_date, planned_harvest_date, status) VALUES (?, ?, ?, 20, '2026-05-10', '2026-08-01', 4)`,
		fieldID, varietyID, seasonID))
	otherPlanID := lastVerifyID(t, execVerifySQL(t, env.DB,
		`INSERT INTO planting_plans (field_id, crop_variety_id, season_id, planned_area, planned_sow_date, planned_harvest_date, status) VALUES (?, ?, ?, 20, '2026-05-11', '2026-08-02', 4)`,
		otherFieldID, varietyID, seasonID))
	return varietyID, planID, otherPlanID
}

func bug031Harvest(t *testing.T, env *verifyEnv, planID int64, date string, weight int) int64 {
	t.Helper()
	return lastVerifyID(t, execVerifySQL(t, env.DB,
		`INSERT INTO harvests (planting_plan_id, harvest_date, total_weight, grade, remark, approved) VALUES (?, ?, ?, 'A', 'BUG031', 0)`,
		planID, date, weight))
}

func assertBug031Approval(t *testing.T, env *verifyEnv, harvestID int64, want int, label string) {
	t.Helper()
	if got := queryVerifyInt(t, env.DB, `SELECT approved FROM harvests WHERE id=?`, harvestID); got != want {
		t.Errorf("%s approved=%d, want %d", label, got, want)
	}
}
