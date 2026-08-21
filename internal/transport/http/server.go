// Package http implements the transport layer: the HTTP server, middleware
// (auth, RBAC, rate limiting, idempotency, structured logging, panic recovery),
// request/response helpers and one handler file per resource.
package http

import (
	"context"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"go.uber.org/zap"

	"go-farm-production/internal/config"
	"go-farm-production/internal/repository"
	"go-farm-production/internal/service"
)

// Server bundles the router, dependencies and runtime config.
type Server struct {
	cfg     *config.Config
	logger  *zap.Logger
	router  *mux.Router
	svc     *service.Container
	store   *repository.Store
	cache   *repository.Cache
	httpSrv *http.Server
}

// New wires the server, applying middleware and mounting every route group.
func New(cfg *config.Config, store *repository.Store, cache *repository.Cache, svc *service.Container, logger *zap.Logger) *Server {
	s := &Server{cfg: cfg, logger: logger, svc: svc, store: store, cache: cache, router: mux.NewRouter()}
	s.router.Use(
		RequestIDMiddleware,
		RecoverMiddleware(logger),
		AccessLogMiddleware(logger),
		BodyLimitMiddleware(cfg.HTTP.MaxBodyBytes),
	)
	s.mountRoutes()
	return s
}

// mountRoutes registers all API endpoints. Public routes skip auth; protected
// routes require a valid JWT + the listed permission(s).
func (s *Server) mountRoutes() {
	api := s.router.PathPrefix("/api/v1").Subrouter()

	// Public (no auth) routes.
	auth := api.PathPrefix("/auth").Subrouter()
	auth.HandleFunc("/register", s.handleRegister).Methods(http.MethodPost)
	auth.HandleFunc("/login", s.handleLogin).Methods(http.MethodPost)
	auth.HandleFunc("/refresh", s.handleRefresh).Methods(http.MethodPost)
	auth.HandleFunc("/logout", s.handleLogout).Methods(http.MethodPost)

	// Health (public, no version prefix).
	s.router.HandleFunc("/health", s.handleHealth).Methods(http.MethodGet)
	s.router.HandleFunc("/ready", s.handleReady).Methods(http.MethodGet)

	// Me: requires auth, no specific permission.
	me := api.PathPrefix("/me").Subrouter()
	me.Use(s.AuthMiddleware)
	me.HandleFunc("", s.handleMe).Methods(http.MethodGet)
	me.HandleFunc("/password", s.handleChangePassword).Methods(http.MethodPatch)

	// All remaining routes require auth.
	protected := api.PathPrefix("").Subrouter()
	protected.Use(s.AuthMiddleware)
	s.mountUsers(protected)
	s.mountRoles(protected)
	s.mountFarms(protected)
	s.mountFields(protected)
	s.mountCropVarieties(protected)
	s.mountSeasons(protected)
	s.mountPlantingPlans(protected)
	s.mountTasks(protected)
	s.mountMaterials(protected)
	s.mountBatches(protected)
	s.mountAllocations(protected)
	s.mountHarvests(protected)
	s.mountProduceInventory(protected)
	s.mountCostAnalyses(protected)
	s.mountReports(protected)
	s.mountAuditLogs(protected)
}

// Run starts the HTTP server.
func (s *Server) Run() error {
	s.httpSrv = &http.Server{
		Addr:         s.cfg.HTTP.Addr,
		Handler:      s.router,
		ReadTimeout:  s.cfg.HTTP.ReadTimeout,
		WriteTimeout: s.cfg.HTTP.WriteTimeout,
		IdleTimeout:  s.cfg.HTTP.IdleTimeout,
	}
	s.logger.Info("http server listening", zap.String("addr", s.cfg.HTTP.Addr))
	return s.httpSrv.ListenAndServe()
}

// Shutdown drains in-flight requests within the configured timeout.
func (s *Server) Shutdown(ctx context.Context) error {
	if s.httpSrv == nil {
		return nil
	}
	return s.httpSrv.Shutdown(ctx)
}

// shutdownTimeout returns the graceful-shutdown deadline.
func (s *Server) shutdownTimeout() time.Duration { return s.cfg.HTTP.ShutdownTimeout }
