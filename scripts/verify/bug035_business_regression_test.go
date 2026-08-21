package verify

import (
	"context"
	"go-farm-production/internal/domain"
	"testing"
	"time"
)

func TestBug035_BusinessRegression(t *testing.T) {
	existing := &domain.FarmTask{PlantingPlanID: 1, TaskType: domain.TaskTypeFertilizing}
	if !domain.TaskAccountingIdentityChanged(existing, &domain.FarmTaskUpsert{PlantingPlanID: 1, TaskType: domain.TaskTypeOther}) {
		t.Error("task type must be protected after accounting history")
	}
	env := newVerifyEnv(t, "BUG-035", false)
	farm := ins35(t, env, `INSERT INTO farms(name,total_area,status) VALUES('f',100,1)`)
	field := ins35(t, env, `INSERT INTO fields(farm_id,code,name,area,status) VALUES(?,'F35','f',100,1)`, farm)
	variety := ins35(t, env, `INSERT INTO crop_varieties(code,name,status) VALUES('V35','v',1)`)
	season := ins35(t, env, `INSERT INTO seasons(code,name,start_date,end_date,status) VALUES('S35','s','2026-01-01','2026-12-31',1)`)
	planA := ins35(t, env, `INSERT INTO planting_plans(field_id,crop_variety_id,season_id,planned_area,planned_sow_date,status,remark) VALUES(?,?,?,?,?,1,'')`, field, variety, season, 50, "2026-01-01")
	planB := ins35(t, env, `INSERT INTO planting_plans(field_id,crop_variety_id,season_id,planned_area,planned_sow_date,status,remark) VALUES(?,?,?,?,?,1,'')`, field, variety, season, 50, "2026-06-01")
	task := ins35(t, env, `INSERT INTO farm_tasks(planting_plan_id,task_type,title,description,planned_date,status,remark) VALUES(?,1,'task','','2026-02-01',1,'')`, planA)
	material := ins35(t, env, `INSERT INTO materials(code,name,category,unit,status) VALUES('M35','m',5,'kg',1)`)
	batch := ins35(t, env, `INSERT INTO input_batches(material_id,batch_no,quantity,remaining_qty,status) VALUES(?,'B35',10,9,1)`, material)
	user := ins35(t, env, `INSERT INTO users(username,email,password_hash,full_name,status) VALUES('u35','u35@test','x','u',1)`)
	if _, err := env.DB.Exec(`UPDATE farm_tasks SET assignee_id=? WHERE id=?`, user, task); err != nil {
		t.Fatal(err)
	}
	ins35(t, env, `INSERT INTO input_allocations(batch_id,material_id,task_id,quantity,type,operator_id) VALUES(?,?,?,?,?,?)`, batch, material, task, 1, domain.AllocationTypeWaste, user)
	u := &domain.FarmTaskUpsert{PlantingPlanID: planB, TaskType: domain.TaskTypeOther, Title: "changed", PlannedDate: "2026-02-01", Status: 1}
	if err := env.Services.Task.Update(context.Background(), task, u); err == nil || domain.AsAppError(err).Code != domain.CodeConflict {
		t.Errorf("service must return conflict for task accounting identity changes: %v", err)
	}
	direct := &domain.FarmTask{PlantingPlanID: planA, TaskType: domain.TaskTypeOther, PlannedDate: time.Now()}
	if err := env.Store.TaskRepo.ValidateAccountingIdentity(context.Background(), env.DB, task, direct); err == nil {
		t.Error("repository must reject task type changes independently")
	}
}
func ins35(t *testing.T, e *verifyEnv, q string, a ...any) int64 {
	r := execVerifySQL(t, e.DB, q, a...)
	id, err := r.LastInsertId()
	if err != nil {
		t.Fatal(err)
	}
	return id
}
