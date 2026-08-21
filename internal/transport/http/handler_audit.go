package http

import (
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"go-farm-production/internal/domain"
	"go-farm-production/internal/repository"
)

// mountAuditLogs registers the audit-log routes (admin + auditor).
func (s *Server) mountAuditLogs(r *mux.Router) {
	g := r.PathPrefix("/audit-logs").Subrouter()
	g.Use(s.Require(domain.PermAuditView))
	g.HandleFunc("", s.listAuditLogs).Methods(http.MethodGet)
}

func (s *Server) listAuditLogs(w http.ResponseWriter, r *http.Request) {
	p := pagination(r)
	f := repository.AuditFilter{
		Action:       r.URL.Query().Get("action"),
		ResourceType: r.URL.Query().Get("resource_type"),
		ResourceID:   r.URL.Query().Get("resource_id"),
		FromTime:     r.URL.Query().Get("from"),
		ToTime:       r.URL.Query().Get("to"),
	}
	if v := r.URL.Query().Get("user_id"); v != "" {
		if id, err := strconv.ParseInt(v, 10, 64); err == nil {
			f.UserID = &id
		}
	}
	rows, total, err := s.store.AuditRepo.List(r.Context(), s.store.DB(), p, f)
	if err != nil {
		writeError(w, err)
		return
	}
	writePage(w, domain.NewPageResult(rows, total, p))
}
