package verify

import (
	"context"
	"testing"
	"time"

	"go-farm-production/internal/domain"
)

func TestBug014_BusinessRegression(t *testing.T) {
	t.Run("unreferenced variety deletes", func(t *testing.T) {
		env := newVerifyEnv(t, "BUG-014-empty", false)
		id := createBug014Variety(t, env, "empty")
		if err := env.Services.Variety.Delete(context.Background(), id); err != nil {
			t.Fatalf("delete unused variety: %v", err)
		}
		assertBug014ExistsStatus(t, env, id, false, 0)
	})
	t.Run("plan reference rejects without changing variety", func(t *testing.T) {
		env := newVerifyEnv(t, "BUG-014-plan", false)
		variety := createBug014Variety(t, env, "plan")
		createBug014Plan(t, env, variety)
		err := env.Services.Variety.Delete(context.Background(), variety)
		assertBug014Conflict(t, err)
		assertBug014ExistsStatus(t, env, variety, true, domain.StatusActive)
	})
	t.Run("inventory reference maps to conflict and leaves status unchanged", func(t *testing.T) {
		env := newVerifyEnv(t, "BUG-014-inventory", false)
		variety := createBug014Variety(t, env, "inventory")
		supporting := createBug014Variety(t, env, "supporting")
		plan := createBug014Plan(t, env, supporting)
		harvest := insertBug014(t, env, `INSERT INTO harvests (planting_plan_id,harvest_date,total_weight,approved) VALUES (?,?,?,1)`, plan, "2026-06-01", 20)
		_ = insertBug014(t, env, `INSERT INTO produce_inventory (crop_variety_id,harvest_id,quantity,status) VALUES (?,?,?,1)`, variety, harvest, 20)
		err := env.Services.Variety.Delete(context.Background(), variety)
		assertBug014Conflict(t, err)
		assertBug014ExistsStatus(t, env, variety, true, domain.StatusActive)
	})
}
func createBug014Variety(t *testing.T, env *verifyEnv, suffix string) int64 {
	t.Helper()
	id, err := env.Services.Variety.Create(context.Background(), &domain.CropVarietyUpsert{Code: "bug014-" + suffix, Name: suffix, GrowthCycle: 90, Status: domain.StatusActive})
	if err != nil {
		t.Fatalf("create variety: %v", err)
	}
	return id
}
func createBug014Plan(t *testing.T, env *verifyEnv, variety int64) int64 {
	t.Helper()
	farm := insertBug014(t, env, `INSERT INTO farms (name,total_area,status) VALUES (?,?,1)`, "f"+time.Now().Format("150405.000000"), 100)
	field := insertBug014(t, env, `INSERT INTO fields (farm_id,code,name,area,status) VALUES (?,?,?,?,1)`, farm, "c"+time.Now().Format("150405.000000"), "f", 10)
	season := insertBug014(t, env, `INSERT INTO seasons (code,name,start_date,end_date,status) VALUES (?,?,?,?,1)`, "s"+time.Now().Format("150405.000000"), "s", "2026-01-01", "2026-12-31")
	return insertBug014(t, env, `INSERT INTO planting_plans (field_id,crop_variety_id,season_id,planned_area,planned_sow_date,status) VALUES (?,?,?,?,?,1)`, field, variety, season, 5, "2026-03-01")
}
func insertBug014(t *testing.T, env *verifyEnv, q string, args ...any) int64 {
	t.Helper()
	r := execVerifySQL(t, env.DB, q, args...)
	id, err := r.LastInsertId()
	if err != nil {
		t.Fatalf("last id: %v", err)
	}
	return id
}
func assertBug014Conflict(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatalf("referenced variety deletion succeeded")
	}
	if ae := domain.AsAppError(err); ae.Code != domain.CodeConflict {
		t.Fatalf("delete code=%d, want %d: %v", ae.Code, domain.CodeConflict, err)
	}
}
func assertBug014ExistsStatus(t *testing.T, env *verifyEnv, id int64, wantExists bool, wantStatus int8) {
	t.Helper()
	var count int
	var status int8
	err := env.DB.QueryRow(`SELECT COUNT(*),COALESCE(MAX(status),0) FROM crop_varieties WHERE id=?`, id).Scan(&count, &status)
	if err != nil {
		t.Fatalf("read variety: %v", err)
	}
	if (count == 1) != wantExists || (wantExists && status != wantStatus) {
		t.Fatalf("variety exists=%v status=%d, want exists=%v status=%d", count == 1, status, wantExists, wantStatus)
	}
}
