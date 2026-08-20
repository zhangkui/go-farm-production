// Package app performs dependency wiring: it builds the logger, opens MySQL
// and Redis connections, runs migrations, seeds the default admin, and
// constructs the service container + HTTP server.
package app

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	_ "github.com/go-sql-driver/mysql" // register the mysql driver for database/sql
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"

	"go-farm-production/internal/config"
	"go-farm-production/internal/domain"
	"go-farm-production/internal/migrate"
	"go-farm-production/internal/repository"
	"go-farm-production/internal/service"
	httptransport "go-farm-production/internal/transport/http"
)

// App holds the long-lived components that main starts and shuts down.
type App struct {
	cfg    *config.Config
	logger *zap.Logger
	db     *sql.DB
	rdb    *redis.Client
	server *httptransport.Server
}

// Bootstrap loads config, opens dependencies, migrates and seeds.
func Bootstrap(ctx context.Context) (*App, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	logger, err := buildLogger(cfg)
	if err != nil {
		return nil, fmt.Errorf("build logger: %w", err)
	}
	logger.Info("starting", zap.String("env", cfg.App.Env), zap.String("name", cfg.App.Name))

	db, err := openMySQL(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("open mysql: %w", err)
	}
	logger.Info("mysql connected", zap.String("dsn", cfg.MySQL.User+"@"+cfg.MySQL.Host))

	rdb := openRedis(cfg)
	if err := pingRedis(ctx, rdb); err != nil {
		logger.Warn("redis ping failed (cache disabled)", zap.Error(err))
	}

	// Migrate + seed.
	runner := migrate.NewRunner(db)
	if ran, err := runner.Up(ctx); err != nil {
		return nil, fmt.Errorf("migrate: %w", err)
	} else if len(ran) > 0 {
		for _, m := range ran {
			logger.Info("migrated", zap.Int("version", m.Version), zap.String("file", m.Name))
		}
	}

	store := repository.New(db, rdb)
	cache := repository.NewCache(rdb)
	hasher := service.NewHasher(cfg.Password.BcryptCost)
	signer := service.NewTokenSigner(cfg.JWT)
	svc := service.New(store, cache, hasher, signer)

	// Seed default admin user on a fresh database.
	if err := seedAdmin(ctx, store, hasher, cfg, logger); err != nil {
		logger.Warn("seed admin", zap.Error(err))
	}

	server := httptransport.New(cfg, store, cache, svc, logger)
	return &App{cfg: cfg, logger: logger, db: db, rdb: rdb, server: server}, nil
}

// Run starts the HTTP server (blocking).
func (a *App) Run() error { return a.server.Run() }

// Shutdown gracefully drains connections and closes DB/Redis.
func (a *App) Shutdown(ctx context.Context) error {
	a.logger.Info("shutting down")
	var first error
	if err := a.server.Shutdown(ctx); err != nil {
		first = err
	}
	_ = a.db.Close()
	if a.rdb != nil {
		_ = a.rdb.Close()
	}
	a.logger.Sync()
	return first
}

// Logger exposes the logger (used by main for fatal logging).
func (a *App) Logger() *zap.Logger { return a.logger }

func buildLogger(cfg *config.Config) (*zap.Logger, error) {
	zcfg := zap.NewProductionConfig()
	zcfg.Encoding = cfg.Log.Format
	switch cfg.Log.Level {
	case "debug":
		zcfg.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	case "info":
		zcfg.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	case "warn":
		zcfg.Level = zap.NewAtomicLevelAt(zap.WarnLevel)
	case "error":
		zcfg.Level = zap.NewAtomicLevelAt(zap.ErrorLevel)
	}
	if !cfg.IsProd() {
		zcfg.Encoding = "console"
		zcfg.Development = true
	}
	return zcfg.Build()
}

func openMySQL(ctx context.Context, cfg *config.Config) (*sql.DB, error) {
	dsn := cfg.MySQL.DSN()
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(cfg.MySQL.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MySQL.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.MySQL.ConnMaxLifetime)
	pingCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err := db.PingContext(pingCtx); err != nil {
		return nil, err
	}
	return db, nil
}

func openRedis(cfg *config.Config) *redis.Client {
	if cfg.Redis.Host == "" {
		return nil
	}
	return redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr(),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
}

func pingRedis(ctx context.Context, rdb *redis.Client) error {
	if rdb == nil {
		return nil
	}
	c, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return rdb.Ping(c).Err()
}

// seedAdmin creates the default administrator account once, when the admin
// username is absent. It assigns the seeded 'admin' role.
func seedAdmin(ctx context.Context, store *repository.Store, hasher service.Hasher, cfg *config.Config, logger *zap.Logger) error {
	if _, err := store.UserRepo.GetByUsername(ctx, store.DB(), cfg.App.DefaultAdmin.Username); err == nil {
		return nil // already exists
	} else {
		// Any non-"invalid credentials" error means a real DB problem: log and skip.
		ae := domain.AsAppError(err)
		if ae.Code != domain.CodeInvalidCreds {
			logger.Warn("seed admin: lookup returned unexpected error", zap.Error(err))
		}
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(cfg.App.DefaultAdmin.Password), cfg.Password.BcryptCost)
	if err != nil {
		return err
	}
	u := &domain.User{
		Username:     cfg.App.DefaultAdmin.Username,
		Email:        cfg.App.DefaultAdmin.Email,
		FullName:     cfg.App.DefaultAdmin.FullName,
		PasswordHash: string(hash),
		Status:       domain.StatusActive,
	}
	uid, err := store.UserRepo.Create(ctx, store.DB(), u)
	if err != nil {
		return err
	}
	role, err := store.RoleRepo.GetByCode(ctx, store.DB(), "admin")
	if err != nil {
		return err
	}
	if err := store.UserRepo.AssignRoles(ctx, store.DB(), uid, []int64{role.ID}); err != nil {
		return err
	}
	logger.Info("default admin seeded", zap.Int64("id", uid), zap.String("username", u.Username))
	return nil
}
