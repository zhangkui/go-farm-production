package http

import (
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"go-farm-production/internal/domain"
	"go-farm-production/internal/service"
)

// mountUsers registers user-management routes (admin-only).
func (s *Server) mountUsers(r *mux.Router) {
	g := r.PathPrefix("/users").Subrouter()
	g.Use(s.Require(domain.PermUserManage))
	g.HandleFunc("", s.listUsers).Methods(http.MethodGet)
	g.HandleFunc("", s.createUser).Methods(http.MethodPost)
	g.HandleFunc("/{id:[0-9]+}", s.getUser).Methods(http.MethodGet)
	g.HandleFunc("/{id:[0-9]+}", s.updateUser).Methods(http.MethodPut)
	g.HandleFunc("/{id:[0-9]+}", s.deleteUser).Methods(http.MethodDelete)
	g.HandleFunc("/{id:[0-9]+}/status", s.updateUserStatus).Methods(http.MethodPatch)
	g.HandleFunc("/{id:[0-9]+}/password", s.resetUserPassword).Methods(http.MethodPost)
	g.HandleFunc("/{id:[0-9]+}/roles", s.assignUserRoles).Methods(http.MethodPut)
}

func (s *Server) listUsers(w http.ResponseWriter, r *http.Request) {
	p := pagination(r)
	res, err := s.svc.User.List(r.Context(), p, p.Search)
	if err != nil {
		writeError(w, err)
		return
	}
	writePage(w, res)
}

func (s *Server) createUser(w http.ResponseWriter, r *http.Request) {
	var u domain.UserUpsert
	if err := parseJSON(r, &u); err != nil {
		writeError(w, domain.Wrap(domain.CodeInvalidJSON, 400, "请求体格式错误", err))
		return
	}
	id, err := s.svc.User.Create(r.Context(), &u)
	if err != nil {
		writeError(w, err)
		return
	}
	writeCreated(w, map[string]any{"id": id})
}

func (s *Server) getUser(w http.ResponseWriter, r *http.Request) {
	id := pathID(r)
	user, roles, err := s.svc.User.Get(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeOK(w, map[string]any{"user": user, "roles": roles})
}

func (s *Server) updateUser(w http.ResponseWriter, r *http.Request) {
	id := pathID(r)
	var u domain.UserUpsert
	if err := parseJSON(r, &u); err != nil {
		writeError(w, domain.Wrap(domain.CodeInvalidJSON, 400, "请求体格式错误", err))
		return
	}
	if err := s.svc.User.Update(r.Context(), id, &u); err != nil {
		writeError(w, err)
		return
	}
	writeOK(w, map[string]any{"id": id})
}

func (s *Server) deleteUser(w http.ResponseWriter, r *http.Request) {
	id := pathID(r)
	if err := s.svc.User.Delete(r.Context(), id); err != nil {
		writeError(w, err)
		return
	}
	writeOK(w, map[string]any{"id": id})
}

type statusRequest struct {
	Status int8 `json:"status"`
}

func (s *Server) updateUserStatus(w http.ResponseWriter, r *http.Request) {
	id := pathID(r)
	var req statusRequest
	if err := parseJSON(r, &req); err != nil {
		writeError(w, domain.Wrap(domain.CodeInvalidJSON, 400, "请求体格式错误", err))
		return
	}
	if err := s.svc.User.UpdateStatus(r.Context(), id, req.Status); err != nil {
		writeError(w, err)
		return
	}
	writeOK(w, map[string]any{"id": id, "status": req.Status})
}

type resetPasswordRequest struct {
	NewPassword string `json:"new_password"`
}

func (s *Server) resetUserPassword(w http.ResponseWriter, r *http.Request) {
	id := pathID(r)
	var req resetPasswordRequest
	if err := parseJSON(r, &req); err != nil {
		writeError(w, domain.Wrap(domain.CodeInvalidJSON, 400, "请求体格式错误", err))
		return
	}
	if err := s.svc.User.ResetPassword(r.Context(), id, req.NewPassword); err != nil {
		writeError(w, err)
		return
	}
	writeOK(w, map[string]any{"id": id})
}

type assignRolesRequest struct {
	RoleIDs []int64 `json:"role_ids"`
}

func (s *Server) assignUserRoles(w http.ResponseWriter, r *http.Request) {
	id := pathID(r)
	var req assignRolesRequest
	if err := parseJSON(r, &req); err != nil {
		writeError(w, domain.Wrap(domain.CodeInvalidJSON, 400, "请求体格式错误", err))
		return
	}
	if err := s.svc.User.AssignRoles(r.Context(), id, req.RoleIDs); err != nil {
		writeError(w, err)
		return
	}
	writeOK(w, map[string]any{"id": id, "role_ids": req.RoleIDs})
}

// pathID extracts the {id} path variable as an int64.
func pathID(r *http.Request) int64 {
	v := mux.Vars(r)["id"]
	n, _ := strconv.ParseInt(v, 10, 64)
	return n
}

// _ keeps the service import referenced for tooling clarity.
var _ = service.RequestMeta{}
