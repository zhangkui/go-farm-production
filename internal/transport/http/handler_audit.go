package http

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"

	"go-farm-production/internal/domain"
)

// mountAuditLogs registers the audit-log routes (admin + auditor).
func (s *Server) mountAuditLogs(r *mux.Router) {
	g := r.PathPrefix("/audit-logs").Subrouter()
	g.Use(s.Require(domain.PermAuditView))
	g.HandleFunc("", s.listAuditLogs).Methods(http.MethodGet)
}

func (s *Server) listAuditLogs(w http.ResponseWriter, r *http.Request) {
	p := pagination(r)
	f := domain.AuditLogQuery{
		Action:       r.URL.Query().Get("action"),
		ResourceType: r.URL.Query().Get("resource_type"),
		ResourceID:   r.URL.Query().Get("resource_id"),
	}
	if v := r.URL.Query().Get("user_id"); v != "" {
		id, err := strconv.ParseInt(v, 10, 64)
		if err != nil || id <= 0 {
			writeError(w, domain.Wrap(domain.CodeValidation, 400, "用户ID必须为正整数", err))
			return
		}
		f.UserID = &id
	}
	for raw, target := range map[string]**time.Time{
		r.URL.Query().Get("from"): &f.FromTime,
		r.URL.Query().Get("to"):   &f.ToTime,
	} {
		if raw == "" {
			continue
		}
		parsed, err := time.Parse(time.RFC3339, raw)
		if err != nil {
			writeError(w, domain.Wrap(domain.CodeValidation, 400, "时间参数格式无效，需符合RFC3339", err))
			return
		}
		*target = &parsed
	}
	result, err := s.svc.Audit.List(r.Context(), p, f)
	if err != nil {
		writeError(w, err)
		return
	}
	writePage(w, result)
}
