package http

import (
	"net/http"

	"go-farm-production/internal/domain"
	"go-farm-production/internal/service"
)

// handleHealth reports liveness.
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeOK(w, map[string]any{"status": "ok"})
}

// handleReady checks DB + Redis connectivity.
func (s *Server) handleReady(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if err := s.store.DB().PingContext(ctx); err != nil {
		writeJSON(w, http.StatusServiceUnavailable, Envelope{Code: 50001, Message: "数据库不可用"})
		return
	}
	if s.cache != nil && s.cache.Available() {
		if err := s.store.RDB().Ping(ctx).Err(); err != nil {
			writeJSON(w, http.StatusServiceUnavailable, Envelope{Code: 50002, Message: "缓存不可用"})
			return
		}
	}
	writeOK(w, map[string]any{"status": "ready"})
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// handleLogin authenticates a user and returns a token pair.
func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := parseJSON(r, &req); err != nil {
		writeError(w, domain.Wrap(domain.CodeInvalidJSON, 400, "请求体格式错误", err))
		return
	}
	r = r.WithContext(service.WithMeta(r.Context(), service.RequestMeta{IPAddress: clientIP(r), UserAgent: r.UserAgent()}))
	pair, err := s.svc.Auth.Login(r.Context(), req.Username, req.Password)
	if err != nil {
		writeError(w, err)
		return
	}
	writeOK(w, pair)
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// handleRefresh rotates a refresh token.
func (s *Server) handleRefresh(w http.ResponseWriter, r *http.Request) {
	var req refreshRequest
	if err := parseJSON(r, &req); err != nil {
		writeError(w, domain.Wrap(domain.CodeInvalidJSON, 400, "请求体格式错误", err))
		return
	}
	r = r.WithContext(service.WithMeta(r.Context(), service.RequestMeta{IPAddress: clientIP(r)}))
	pair, err := s.svc.Auth.Refresh(r.Context(), req.RefreshToken)
	if err != nil {
		writeError(w, err)
		return
	}
	writeOK(w, pair)
}

// handleLogout revokes the supplied refresh token.
func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	var req refreshRequest
	_ = parseJSON(r, &req)
	if req.RefreshToken == "" {
		q := r.URL.Query().Get("refresh_token")
		req.RefreshToken = q
	}
	_ = s.svc.Auth.Logout(r.Context(), req.RefreshToken)
	writeOK(w, map[string]any{"status": "ok"})
}

// handleRegister is the public self-registration endpoint (creates a regular
// operator-scoped user).
func (s *Server) handleRegister(w http.ResponseWriter, r *http.Request) {
	var u domain.UserUpsert
	if err := parseJSON(r, &u); err != nil {
		writeError(w, domain.Wrap(domain.CodeInvalidJSON, 400, "请求体格式错误", err))
		return
	}
	r = r.WithContext(service.WithMeta(r.Context(), service.RequestMeta{IPAddress: clientIP(r), UserAgent: r.UserAgent()}))
	id, err := s.svc.User.Create(r.Context(), &u)
	if err != nil {
		writeError(w, err)
		return
	}
	writeCreated(w, map[string]any{"id": id})
}

// handleMe returns the current user's profile and permissions.
func (s *Server) handleMe(w http.ResponseWriter, r *http.Request) {
	meta, _ := service.MetaFrom(r.Context())
	user, roles, err := s.svc.User.Get(r.Context(), meta.UserID)
	if err != nil {
		writeError(w, err)
		return
	}
	_, codes, err := s.svc.Auth.Identity(r.Context(), meta.UserID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeOK(w, map[string]any{
		"user":        user,
		"roles":       roles,
		"permissions": codes,
	})
}

type changePasswordRequest struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

// handleChangePassword lets an authenticated user change their own password.
func (s *Server) handleChangePassword(w http.ResponseWriter, r *http.Request) {
	var req changePasswordRequest
	if err := parseJSON(r, &req); err != nil {
		writeError(w, domain.Wrap(domain.CodeInvalidJSON, 400, "请求体格式错误", err))
		return
	}
	meta, _ := service.MetaFrom(r.Context())
	if err := s.svc.User.UpdatePassword(r.Context(), meta.UserID, req.OldPassword, req.NewPassword); err != nil {
		writeError(w, err)
		return
	}
	writeOK(w, map[string]any{"status": "ok"})
}
