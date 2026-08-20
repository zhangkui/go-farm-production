package verify

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/go-sql-driver/mysql"

	"go-farm-production/internal/domain"
	"go-farm-production/internal/repository"
	"go-farm-production/internal/service"
)

type noOpAudit struct{}

func (noOpAudit) Log(context.Context, domain.AuditEntry) error { return nil }

type fixture struct {
	db      *sql.DB
	service service.InputAllocationService
	batchID int64
}

func TestBug001_BusinessRegression(t *testing.T) {
	taskOne := int64(101)
	taskTwo := int64(202)

	t.Run("valid partial and exact returns", func(t *testing.T) {
		f := newFixture(t, 100)
		mustMove(t, f.service, f.batchID, &taskOne, 60, domain.AllocationTypeAllocate)
		mustMove(t, f.service, f.batchID, &taskOne, 20, domain.AllocationTypeReturn)
		mustMove(t, f.service, f.batchID, &taskOne, 40, domain.AllocationTypeReturn)
		assertStock(t, f.db, f.batchID, 100)
		assertMovementCount(t, f.db, f.batchID, 3)
	})

	t.Run("cross task return is rejected and rolled back", func(t *testing.T) {
		f := newFixture(t, 100)
		mustMove(t, f.service, f.batchID, &taskOne, 60, domain.AllocationTypeAllocate)
		mustMove(t, f.service, f.batchID, &taskTwo, 20, domain.AllocationTypeAllocate)
		before := movementCount(t, f.db, f.batchID)
		assertRejected(t, move(f.service, f.batchID, &taskTwo, 30, domain.AllocationTypeReturn))
		assertStock(t, f.db, f.batchID, 20)
		assertMovementCount(t, f.db, f.batchID, before)
	})

	t.Run("waste is irreversible", func(t *testing.T) {
		f := newFixture(t, 100)
		mustMove(t, f.service, f.batchID, &taskOne, 30, domain.AllocationTypeWaste)
		assertRejected(t, move(f.service, f.batchID, &taskOne, 1, domain.AllocationTypeReturn))
		assertStock(t, f.db, f.batchID, 70)
		assertMovementCount(t, f.db, f.batchID, 1)
	})

	t.Run("cumulative return cannot exceed allocation", func(t *testing.T) {
		f := newFixture(t, 100)
		mustMove(t, f.service, f.batchID, &taskOne, 25, domain.AllocationTypeAllocate)
		mustMove(t, f.service, f.batchID, &taskOne, 10, domain.AllocationTypeReturn)
		mustMove(t, f.service, f.batchID, &taskOne, 15, domain.AllocationTypeReturn)
		assertRejected(t, move(f.service, f.batchID, &taskOne, 0.01, domain.AllocationTypeReturn))
		assertStock(t, f.db, f.batchID, 100)
		assertMovementCount(t, f.db, f.batchID, 3)
	})

	t.Run("nil task remains isolated", func(t *testing.T) {
		f := newFixture(t, 100)
		mustMove(t, f.service, f.batchID, nil, 12, domain.AllocationTypeAllocate)
		mustMove(t, f.service, f.batchID, nil, 12, domain.AllocationTypeReturn)
		assertRejected(t, move(f.service, f.batchID, &taskOne, 1, domain.AllocationTypeReturn))
		assertStock(t, f.db, f.batchID, 100)
	})

	t.Run("two concurrent returns serialize", func(t *testing.T) {
		f := newFixture(t, 100)
		taskID := int64(303)
		mustMove(t, f.service, f.batchID, &taskID, 40, domain.AllocationTypeAllocate)

		start := make(chan struct{})
		errs := make(chan error, 2)
		var workers sync.WaitGroup
		for range 2 {
			workers.Add(1)
			go func() {
				defer workers.Done()
				<-start
				errs <- move(f.service, f.batchID, &taskID, 30, domain.AllocationTypeReturn)
			}()
		}
		close(start)
		workers.Wait()
		close(errs)

		var successes, rejected int
		for err := range errs {
			switch {
			case err == nil:
				successes++
			case errors.Is(err, domain.ErrReturnExceedsUsed):
				rejected++
			default:
				t.Fatalf("unexpected concurrent error: %v", err)
			}
		}
		if successes != 1 || rejected != 1 {
			t.Fatalf("successes=%d rejected=%d, want 1/1", successes, rejected)
		}
		assertStock(t, f.db, f.batchID, 90)
		assertMovementCount(t, f.db, f.batchID, 2)
	})
}

func newFixture(t *testing.T, quantity float64) *fixture {
	t.Helper()
	adminDSN := os.Getenv("TEST_MYSQL_DSN")
	if adminDSN == "" {
		adminDSN = "root:rootsecret@tcp(127.0.0.1:3306)/?parseTime=true&multiStatements=true"
	}
	cfg, err := mysql.ParseDSN(adminDSN)
	if err != nil {
		t.Fatalf("parse TEST_MYSQL_DSN: %v", err)
	}
	cfg.DBName = ""
	adminDB, err := sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		t.Fatalf("open mysql: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := adminDB.PingContext(ctx); err != nil {
		_ = adminDB.Close()
		t.Fatalf("mysql is required: %v", err)
	}
	dbName := fmt.Sprintf("goxm_bug001_%d", time.Now().UnixNano())
	if _, err := adminDB.ExecContext(ctx, "CREATE DATABASE `"+dbName+"` CHARACTER SET utf8mb4"); err != nil {
		_ = adminDB.Close()
		t.Fatalf("create database: %v", err)
	}
	t.Cleanup(func() {
		_, _ = adminDB.Exec("DROP DATABASE IF EXISTS `" + dbName + "`")
		_ = adminDB.Close()
	})
	cfg.DBName = dbName
	db, err := sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	db.SetMaxOpenConns(8)
	t.Cleanup(func() { _ = db.Close() })

	statements := []string{
		`CREATE TABLE materials (id BIGINT NOT NULL PRIMARY KEY, name VARCHAR(128) NOT NULL) ENGINE=InnoDB`,
		`CREATE TABLE input_batches (
			id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
			material_id BIGINT NOT NULL,
			batch_no VARCHAR(64) NOT NULL,
			quantity DECIMAL(12,2) NOT NULL,
			remaining_qty DECIMAL(12,2) NOT NULL,
			purchase_date DATE NULL,
			expiry_date DATE NULL,
			purchase_price DECIMAL(12,2) NOT NULL DEFAULT 0,
			supplier VARCHAR(128) NOT NULL DEFAULT '',
			status TINYINT NOT NULL,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
		) ENGINE=InnoDB`,
		`CREATE TABLE input_allocations (
			id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
			batch_id BIGINT NOT NULL,
			material_id BIGINT NOT NULL,
			task_id BIGINT NULL,
			quantity DECIMAL(12,2) NOT NULL,
			type TINYINT NOT NULL,
			remark TEXT,
			operator_id BIGINT NOT NULL DEFAULT 0,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			KEY idx_returnable (batch_id, task_id, type)
		) ENGINE=InnoDB`,
	}
	for _, statement := range statements {
		if _, err := db.ExecContext(ctx, statement); err != nil {
			t.Fatalf("create schema: %v", err)
		}
	}
	if _, err := db.ExecContext(ctx, `INSERT INTO materials (id, name) VALUES (1, '验证物料')`); err != nil {
		t.Fatalf("insert material: %v", err)
	}
	result, err := db.ExecContext(ctx, `INSERT INTO input_batches
		(material_id, batch_no, quantity, remaining_qty, status) VALUES (1, 'VERIFY-001', ?, ?, ?)`,
		quantity, quantity, domain.BatchStatusActive)
	if err != nil {
		t.Fatalf("insert batch: %v", err)
	}
	batchID, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("batch id: %v", err)
	}
	store := repository.New(db, nil)
	return &fixture{db: db, service: service.NewInputAllocationService(store, noOpAudit{}), batchID: batchID}
}

func move(allocationService service.InputAllocationService, batchID int64, taskID *int64, quantity float64, allocationType int8) error {
	_, err := allocationService.Allocate(service.WithMeta(context.Background(), service.RequestMeta{UserID: 1, Username: "verification"}), &domain.InputAllocationUpsert{
		BatchID: batchID, TaskID: taskID, Quantity: quantity, Type: allocationType,
	})
	return err
}

func mustMove(t *testing.T, allocationService service.InputAllocationService, batchID int64, taskID *int64, quantity float64, allocationType int8) {
	t.Helper()
	if err := move(allocationService, batchID, taskID, quantity, allocationType); err != nil {
		t.Fatalf("movement type=%d quantity=%.2f failed: %v", allocationType, quantity, err)
	}
}

func assertRejected(t *testing.T, err error) {
	t.Helper()
	if !errors.Is(err, domain.ErrReturnExceedsUsed) {
		t.Fatalf("return error=%v, want ErrReturnExceedsUsed", err)
	}
}

func assertStock(t *testing.T, db *sql.DB, batchID int64, want float64) {
	t.Helper()
	var got float64
	if err := db.QueryRow(`SELECT remaining_qty FROM input_batches WHERE id=?`, batchID).Scan(&got); err != nil {
		t.Fatalf("read stock: %v", err)
	}
	if got != want {
		t.Fatalf("remaining stock=%.2f want %.2f", got, want)
	}
}

func movementCount(t *testing.T, db *sql.DB, batchID int64) int {
	t.Helper()
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM input_allocations WHERE batch_id=?`, batchID).Scan(&count); err != nil {
		t.Fatalf("count movements: %v", err)
	}
	return count
}

func assertMovementCount(t *testing.T, db *sql.DB, batchID int64, want int) {
	t.Helper()
	if got := movementCount(t, db, batchID); got != want {
		t.Fatalf("movement rows=%d want %d", got, want)
	}
}
