package verify

import (
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"go-farm-production/internal/domain"
)

var lateBugSequence atomic.Int64

func lateBugCode(prefix string) string { return fmt.Sprintf("%s-%d", prefix, lateBugSequence.Add(1)) }

func lateInsertID(t *testing.T, env *verifyEnv, query string, args ...any) int64 {
	t.Helper()
	result := execVerifySQL(t, env.DB, query, args...)
	id, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("last insert id: %v", err)
	}
	return id
}

func latePlanFixture(t *testing.T, env *verifyEnv, status int8) (int64, int64, int64) {
	t.Helper()
	code := lateBugCode("late")
	farmID := lateInsertID(t, env, `INSERT INTO farms (name,total_area,status) VALUES (?,?,1)`, code, 100)
	fieldID := lateInsertID(t, env, `INSERT INTO fields (farm_id,code,name,area,status,remark) VALUES (?,?,?,?,1,?)`, farmID, code, code, 50, "fixture")
	varietyID := lateInsertID(t, env, `INSERT INTO crop_varieties (code,name,category,growth_cycle,description,status) VALUES (?,?,?,?,?,1)`, code, code, "fixture", 90, "fixture")
	seasonID := lateInsertID(t, env, `INSERT INTO seasons (code,name,start_date,end_date,status) VALUES (?,?,?,?,1)`, code, code, "2026-01-01", "2026-12-31")
	planID := lateInsertID(t, env, `INSERT INTO planting_plans (field_id,crop_variety_id,season_id,planned_area,planned_sow_date,planned_harvest_date,status,remark) VALUES (?,?,?,?,?,?,?,?)`, fieldID, varietyID, seasonID, 20, "2026-03-01", "2026-06-01", status, "fixture")
	return planID, fieldID, varietyID
}

func lateTaskFixture(t *testing.T, env *verifyEnv, planID int64, status int8) int64 {
	t.Helper()
	code := lateBugCode("task")
	assigneeID := lateInsertID(t, env, `INSERT INTO users (username,email,full_name,password_hash,status) VALUES (?,?,?,?,1)`, code, code+"@test", code, "hash")
	return lateInsertID(t, env, `INSERT INTO farm_tasks (planting_plan_id,title,description,planned_date,status,assignee_id,remark) VALUES (?,?,?,?,?,?,?)`, planID, code, "fixture", "2026-04-01", status, assigneeID, "fixture")
}

func lateAllocationFixture(t *testing.T, env *verifyEnv, taskID int64) {
	t.Helper()
	code := lateBugCode("alloc")
	userID := lateInsertID(t, env, `INSERT INTO users (username,email,full_name,password_hash,status) VALUES (?,?,?,?,1)`, code, code+"@test", code, "hash")
	materialID := lateInsertID(t, env, `INSERT INTO materials (code,name,category,unit,description,status) VALUES (?,?,?,?,?,1)`, code, code, 1, "kg", "fixture")
	batchID := lateInsertID(t, env, `INSERT INTO input_batches (material_id,batch_no,quantity,remaining_qty,purchase_date,expiry_date,supplier,status) VALUES (?,?,?,?,?,?,?,1)`, materialID, code, 10, 5, "2026-01-01", "2027-01-01", "fixture")
	lateInsertID(t, env, `INSERT INTO input_allocations (batch_id,material_id,task_id,quantity,type,remark,operator_id) VALUES (?,?,?,?,?,?,?)`, batchID, materialID, taskID, 5, domain.AllocationTypeAllocate, "fixture", userID)
}

func lateHarvestFixture(t *testing.T, env *verifyEnv, planID int64, approved bool) int64 {
	t.Helper()
	return lateInsertID(t, env, `INSERT INTO harvests (planting_plan_id,harvest_date,total_weight,grade,remark,approved) VALUES (?,?,?,?,?,?)`, planID, time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC), 10, "A", "fixture", approved)
}
