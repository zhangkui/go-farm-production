package http

import (
	"context"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"go.uber.org/zap"

	"go-farm-production/internal/repository"
	"go-farm-production/internal/service"
)

// ctxKey is an unexported type for context keys in this package.
type ctxKey string

const (
	ctxReqID ctxKey = "req_id"
	ctxMeta  ctxKey = "meta"
)

// RequestIDMiddleware injects a unique request id into the context and the
// X-Request-Id response header.
func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-Id")
		if id == "" {
			id = newReqID()
		}
		w.Header().Set("X-Request-Id", id)
		ctx := context.WithValue(r.Context(), ctxReqID, id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RecoverMiddleware turns panics into 500s without crashing the process.
func RecoverMiddleware(logger *zap.Logger) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					logger.Error("panic recovered",
						zap.Any("recover", rec),
						zap.String("stack", string(debug.Stack())),
						zap.String("path", r.URL.Path),
					)
					writeJSON(w, http.StatusInternalServerError, Envelope{
						Code:    50000,
						Message: "服务器内部错误",
					})
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

// AccessLogMiddleware logs each request as a structured JSON line.
func AccessLogMiddleware(logger *zap.Logger) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			ww := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(ww, r)
			logger.Info("http",
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.Int("status", ww.status),
				zap.Duration("latency", time.Since(start)),
				zap.String("ip", clientIP(r)),
			)
		})
	}
}

// BodyLimitMiddleware rejects oversized request bodies.
func BodyLimitMiddleware(max int64) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.ContentLength > max {
				writeJSON(w, http.StatusRequestEntityTooLarge, Envelope{
					Code: 40000, Message: "请求体过大",
				})
				return
			}
			r.Body = http.MaxBytesReader(w, r.Body, max)
			next.ServeHTTP(w, r)
		})
	}
}

// AuthMiddleware validates the JWT access token and injects the authenticated
// principal's RequestMeta into the request context.
func (s *Server) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := extractBearer(r)
		if token == "" {
			writeJSON(w, http.StatusUnauthorized, Envelope{Code: 40100, Message: "未认证"})
			return
		}
		claims, err := s.svc.Tokens.ParseAccessToken(r.Context(), token)
		if err != nil {
			writeError(w, err)
			return
		}
		// Load username/permissions for the principal.
		ident, _, err := s.svc.Auth.Identity(r.Context(), claims.UserID)
		if err != nil {
			writeError(w, err)
			return
		}
		meta := service.RequestMeta{
			UserID:    ident.UserID,
			Username:  ident.Username,
			IPAddress: clientIP(r),
			UserAgent: r.UserAgent(),
		}
		ctx := service.WithMeta(r.Context(), meta)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Require returns a middleware that checks the principal holds a permission.
// It must run after AuthMiddleware.
func (s *Server) Require(perm string) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			meta, ok := service.MetaFrom(r.Context())
			if !ok {
				writeJSON(w, http.StatusUnauthorized, Envelope{Code: 40100, Message: "未认证"})
				return
			}
			codes, err := s.loadPerms(r.Context(), meta.UserID)
			if err != nil {
				writeError(w, err)
				return
			}
			if !hasPerm(codes, perm) {
				writeJSON(w, http.StatusForbidden, Envelope{Code: 40301, Message: "无操作权限"})
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// loadPerms resolves the user's permission codes, preferring the cache.
func (s *Server) loadPerms(ctx context.Context, userID int64) ([]string, error) {
	_, codes, err := s.svc.Auth.Identity(ctx, userID)
	return codes, err
}

// hasPerm reports whether codes contains perm (or the wildcard admin scope).
func hasPerm(codes []string, perm string) bool {
	for _, c := range codes {
		if c == perm {
			return true
		}
	}
	return false
}

// extractBearer pulls the token out of an Authorization: Bearer <t> header.
func extractBearer(r *http.Request) string {
	h := r.Header.Get("Authorization")
	if h == "" {
		return ""
	}
	parts := strings.SplitN(h, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
		return ""
	}
	return strings.TrimSpace(parts[1])
}

// clientIP extracts the real client IP, honouring X-Forwarded-For.
func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		if i := strings.Index(xff, ","); i > 0 {
			return strings.TrimSpace(xff[:i])
		}
		return strings.TrimSpace(xff)
	}
	// strip port
	addr := r.RemoteAddr
	if i := strings.LastIndex(addr, ":"); i > 0 && !strings.Contains(addr[i:], "]") {
		return addr[:i]
	}
	return addr
}

// statusRecorder captures the response status for access logging.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (s *statusRecorder) WriteHeader(code int) {
	s.status = code
	s.ResponseWriter.WriteHeader(code)
}

// IdempotencyMiddleware guards write endpoints with an Idempotency-Key. On first
// receipt it caches the response for a short TTL; a duplicate key within the
// TTL replays the cached response instead of re-executing the operation.
func (s *Server) IdempotencyMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.Header.Get("Idempotency-Key")
		if key == "" || s.cache == nil || !s.cache.Available() {
			next.ServeHTTP(w, r)
			return
		}
		ck := repository.IdemKey(key)
		acquired, err := s.cache.SetNX(r.Context(), ck, "1", 10*time.Minute)
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}
		if !acquired {
			writeJSON(w, http.StatusConflict, Envelope{Code: 40902, Message: "幂等键重复，请勿重复提交"})
			return
		}
		next.ServeHTTP(w, r)
	})
}
