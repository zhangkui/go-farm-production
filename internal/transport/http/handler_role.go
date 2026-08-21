package http

import (
	"net/http"

	"github.com/gorilla/mux"

	"go-farm-production/internal/domain"
)

// mountRoles registers role & permission routes (admin-only).
func (s *Server) mountRoles(r *mux.Router) {
	g := r.PathPrefix("/roles").Subrouter()
	g.Use(s.Require(domain.PermRoleManage))
	g.HandleFunc("", s.listRoles).Methods(http.MethodGet)
	g.HandleFunc("", s.createRole).Methods(http.MethodPost)
	g.HandleFunc("/{id:[0-9]+}", s.getRole).Methods(http.MethodGet)
	g.HandleFunc("/{id:[0-9]+}", s.updateRole).Methods(http.MethodPut)
	g.HandleFunc("/{id:[0-9]+}", s.deleteRole).Methods(http.MethodDelete)
	g.HandleFunc("/{id:[0-9]+}/permissions", s.assignRolePermissions).Methods(http.MethodPut)

	pg := r.PathPrefix("/permissions").Subrouter()
	pg.Use(s.Require(domain.PermRoleView))
	pg.HandleFunc("", s.listPermissions).Methods(http.MethodGet)
}

func (s *Server) listRoles(w http.ResponseWriter, r *http.Request) {
	p := pagination(r)
	res, err := s.svc.Role.List(r.Context(), p, p.Search)
	if err != nil {
		writeError(w, err)
		return
	}
	writePage(w, res)
}

func (s *Server) createRole(w http.ResponseWriter, r *http.Request) {
	var u domain.RoleUpsert
	if err := parseJSON(r, &u); err != nil {
		writeError(w, domain.Wrap(domain.CodeInvalidJSON, 400, "请求体格式错误", err))
		return
	}
	id, err := s.svc.Role.Create(r.Context(), &u)
	if err != nil {
		writeError(w, err)
		return
	}
	writeCreated(w, map[string]any{"id": id})
}

func (s *Server) getRole(w http.ResponseWriter, r *http.Request) {
	id := pathID(r)
	role, perms, err := s.svc.Role.Get(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeOK(w, map[string]any{"role": role, "permissions": perms})
}

func (s *Server) updateRole(w http.ResponseWriter, r *http.Request) {
	id := pathID(r)
	var u domain.RoleUpsert
	if err := parseJSON(r, &u); err != nil {
		writeError(w, domain.Wrap(domain.CodeInvalidJSON, 400, "请求体格式错误", err))
		return
	}
	if err := s.svc.Role.Update(r.Context(), id, &u); err != nil {
		writeError(w, err)
		return
	}
	writeOK(w, map[string]any{"id": id})
}

func (s *Server) deleteRole(w http.ResponseWriter, r *http.Request) {
	id := pathID(r)
	if err := s.svc.Role.Delete(r.Context(), id); err != nil {
		writeError(w, err)
		return
	}
	writeOK(w, map[string]any{"id": id})
}

type assignPermissionsRequest struct {
	PermissionIDs []int64 `json:"permission_ids"`
}

func (s *Server) assignRolePermissions(w http.ResponseWriter, r *http.Request) {
	id := pathID(r)
	var req assignPermissionsRequest
	if err := parseJSON(r, &req); err != nil {
		writeError(w, domain.Wrap(domain.CodeInvalidJSON, 400, "请求体格式错误", err))
		return
	}
	if err := s.svc.Role.AssignPermissions(r.Context(), id, req.PermissionIDs); err != nil {
		writeError(w, err)
		return
	}
	writeOK(w, map[string]any{"id": id, "permission_ids": req.PermissionIDs})
}

func (s *Server) listPermissions(w http.ResponseWriter, r *http.Request) {
	perms, err := s.svc.Role.ListPermissions(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	writeOK(w, map[string]any{"items": perms})
}
