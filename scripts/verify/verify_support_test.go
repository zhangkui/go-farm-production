package verify

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/go-sql-driver/mysql"

	"go-farm-production/internal/config"
	"go-farm-production/internal/migrate"
	"go-farm-production/internal/repository"
	"go-farm-production/internal/service"
)

type verifyEnv struct {
	DB       *sql.DB
	Store    *repository.Store
	Services *service.Container
}

func newVerifyEnv(t *testing.T, bugID string) *verifyEnv {
	t.Helper()
	adminDSN := os.Getenv("TEST_MYSQL_DSN")
	if adminDSN == "" {
		adminDSN = "root:rootsecret@tcp(127.0.0.1:3306)/?parseTime=true&multiStatements=true&charset=utf8mb4"
	}
	cfg, err := mysql.ParseDSN(adminDSN)
	if err != nil {
		t.Fatalf("parse TEST_MYSQL_DSN: %v", err)
	}
	cfg.DBName = ""
	adminDB, err := sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		t.Fatalf("open mysql admin connection: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	if err := adminDB.PingContext(ctx); err != nil {
		_ = adminDB.Close()
		t.Fatalf("mysql is required: %v", err)
	}
	dbName := fmt.Sprintf("goxm_%s_%d", strings.ToLower(strings.ReplaceAll(bugID, "-", "")), time.Now().UnixNano())
	if _, err := adminDB.ExecContext(ctx, "CREATE DATABASE `"+dbName+"` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci"); err != nil {
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
	db.SetMaxOpenConns(20)
	t.Cleanup(func() { _ = db.Close() })
	if _, err := migrate.NewRunner(db).Up(ctx); err != nil {
		t.Fatalf("run migrations: %v", err)
	}
	store := repository.New(db, nil)
	cache := repository.NewCache(nil)
	jwtConfig := config.JWTConfig{
		Secret:          strings.Repeat("v", 40),
		Issuer:          "goxm-verification",
		AccessTokenTTL:  time.Hour,
		RefreshTokenTTL: 24 * time.Hour,
	}
	services := service.New(store, cache, service.NewHasher(4), service.NewTokenSigner(jwtConfig))
	return &verifyEnv{DB: db, Store: store, Services: services}
}

func queryVerifyInt(t *testing.T, db *sql.DB, query string, args ...any) int {
	t.Helper()
	var value int
	if err := db.QueryRowContext(context.Background(), query, args...).Scan(&value); err != nil {
		t.Fatalf("query verification integer: %v", err)
	}
	return value
}
