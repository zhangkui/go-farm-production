// Package repository implements the data-access layer. It owns the MySQL
// connection pool, a Redis client, and a per-entity repository. Services
// depend on the Store, which can run multi-table operations inside a single
// transaction via WithTx.
package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/redis/go-redis/v9"

	"go-farm-production/internal/domain"
)

// Store is the single data-access façade handed to the service layer. It
// exposes per-entity repositories and a transaction helper. All repositories
// operate against the same *sql.DB, or a *sql.Tx when inside WithTx.
type Store struct {
	db  domain.TxStarter
	rdb *redis.Client

	UserRepo     UserRepository
	RoleRepo     RoleRepository
	PermRepo     PermissionRepository
	TokenRepo    RefreshTokenRepository
	FarmRepo     FarmRepository
	FieldRepo    FieldRepository
	VarietyRepo  CropVarietyRepository
	SeasonRepo   SeasonRepository
	PlanRepo     PlantingPlanRepository
	TaskRepo     FarmTaskRepository
	MaterialRepo MaterialRepository
	BatchRepo    InputBatchRepository
	AllocRepo    InputAllocationRepository
	HarvestRepo  HarvestRepository
	HDetailRepo  HarvestDetailRepository
	ProduceRepo  ProduceInventoryRepository
	CostRepo     CostAnalysisRepository
	AuditRepo    AuditLogRepository
}

// New opens nothing itself but wraps an already-connected *sql.DB and Redis
// client, wiring every repository to them.
func New(db *sql.DB, rdb *redis.Client) *Store {
	s := &Store{db: db, rdb: rdb}
	s.UserRepo = NewUserRepository(db)
	s.RoleRepo = NewRoleRepository(db)
	s.PermRepo = NewPermissionRepository(db)
	s.TokenRepo = NewRefreshTokenRepository(db)
	s.FarmRepo = NewFarmRepository(db)
	s.FieldRepo = NewFieldRepository(db)
	s.VarietyRepo = NewCropVarietyRepository(db)
	s.SeasonRepo = NewSeasonRepository(db)
	s.PlanRepo = NewPlantingPlanRepository(db)
	s.TaskRepo = NewFarmTaskRepository(db)
	s.MaterialRepo = NewMaterialRepository(db)
	s.BatchRepo = NewInputBatchRepository(db)
	s.AllocRepo = NewInputAllocationRepository(db)
	s.HarvestRepo = NewHarvestRepository(db)
	s.HDetailRepo = NewHarvestDetailRepository(db)
	s.ProduceRepo = NewProduceInventoryRepository(db)
	s.CostRepo = NewCostAnalysisRepository(db)
	s.AuditRepo = NewAuditLogRepository(db)
	return s
}

// DB returns the underlying connection pool (used by migrate).
func (s *Store) DB() *sql.DB { return s.db.(*sql.DB) }

// RDB returns the Redis client.
func (s *Store) RDB() *redis.Client { return s.rdb }

// WithTx runs fn inside a single DB transaction. If fn returns an error the
// transaction is rolled back; otherwise it is committed. Panics are recovered
// and converted to a rollback + error so a single bad statement cannot poison
// the connection.
func (s *Store) WithTx(ctx context.Context, fn func(ctx context.Context, tx domain.DBTX) error) (err error) {
	tx, beginErr := s.db.BeginTx(ctx, nil)
	if beginErr != nil {
		return domain.Wrap(domain.CodeTx, 500, "开启事务失败", beginErr)
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			err = fmt.Errorf("transaction panicked: %v", p)
			return
		}
		if err != nil {
			_ = tx.Rollback()
			return
		}
		err = tx.Commit()
	}()
	return fn(ctx, tx)
}

// BaseRepo embeds a DBTX so each concrete repository can be constructed either
// against the pool or against a transaction. Concrete repositories store a
// domain.DBTX (which *sql.DB also satisfies).
type BaseRepo struct {
	db domain.DBTX
}

// Txable returns whether the repo runs against a live transaction. Always
// false for the pool-backed store repositories; the service-layer WithTx path
// re-implements its calls against the tx directly.
func (BaseRepo) Txable() bool { return false }
