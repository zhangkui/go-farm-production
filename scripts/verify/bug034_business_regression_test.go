package verify

import (
	"context"
	"testing"

	"go-farm-production/internal/domain"
)

func TestBug034_BusinessRegression(t *testing.T) {
	remaining, err := domain.RecalculateBatchRemaining(100, 30, 5, 10)
	if err != nil || remaining != 65 {
		t.Errorf("ledger calculation ignored irreversible waste: remaining=%v err=%v", remaining, err)
	}
	env := newVerifyEnv(t, "BUG-034", false)
	material := insert34(t, env, `INSERT INTO materials(code,name,category,unit,status) VALUES('M34','m',5,'kg',1)`)
	batch := insert34(t, env, `INSERT INTO input_batches(material_id,batch_no,quantity,remaining_qty,status) VALUES(?,'B34',100,65,1)`, material)
	user := insert34(t, env, `INSERT INTO users(username,email,password_hash,full_name,status) VALUES('u34','u34@test','x','u',1)`)
	insert34(t, env, `INSERT INTO input_allocations(batch_id,material_id,quantity,type,operator_id) VALUES(?,?,?,?,?)`, batch, material, 30, domain.AllocationTypeAllocate, user)
	insert34(t, env, `INSERT INTO input_allocations(batch_id,material_id,quantity,type,operator_id) VALUES(?,?,?,?,?)`, batch, material, 10, domain.AllocationTypeWaste, user)
	insert34(t, env, `INSERT INTO input_allocations(batch_id,material_id,quantity,type,operator_id) VALUES(?,?,?,?,?)`, batch, material, 5, domain.AllocationTypeReturn, user)
	update := &domain.InputBatchUpsert{MaterialID: material, BatchNo: "B34", Quantity: 20, PurchaseDate: "2026-01-01", ExpiryDate: "2027-01-01", Status: domain.BatchStatusActive}
	if err := env.Services.Batch.Update(context.Background(), batch, update); err == nil || domain.AsAppError(err).Code != domain.CodeConflict {
		t.Errorf("service must return conflict below irreversible net consumption: %v", err)
	}
	var quantity, storedRemaining float64
	if err := env.DB.QueryRow(`SELECT quantity,remaining_qty FROM input_batches WHERE id=?`, batch).Scan(&quantity, &storedRemaining); err != nil || quantity != 100 || storedRemaining != 65 {
		t.Fatalf("rejected correction changed batch: quantity=%v remaining=%v err=%v", quantity, storedRemaining, err)
	}
	if err := env.Store.BatchRepo.UpdateQuantityGuarded(context.Background(), env.Store.DB(), batch, update, -15); err == nil {
		t.Error("repository must independently reject impossible cached balances")
	}
}

func insert34(t *testing.T, env *verifyEnv, query string, args ...any) int64 {
	t.Helper()
	result := execVerifySQL(t, env.DB, query, args...)
	id, err := result.LastInsertId()
	if err != nil {
		t.Fatal(err)
	}
	return id
}
