package http

import (
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"go-farm-production/internal/domain"
	"go-farm-production/internal/repository"
)

// mountAllocations registers input-allocation routes.
func (s *Server) mountAllocations(r *mux.Router) {
	g := r.PathPrefix("/allocations").Subrouter()
	g.Use(s.Require(domain.PermAllocationManage))
	g.Handle("", s.IdempotencyMiddleware(http.HandlerFunc(s.createAllocation))).Methods(http.MethodPost)
	g.Handle("", http.HandlerFunc(s.listAllocations)).Methods(http.MethodGet)
	g.HandleFunc("/{id:[0-9]+}", s.getAllocation).Methods(http.MethodGet)
}

func (s *Server) createAllocation(w http.ResponseWriter, r *http.Request) {
	var u domain.InputAllocationUpsert
	if err := parseJSON(r, &u); err != nil {
		writeError(w, domain.Wrap(domain.CodeInvalidJSON, 400, "请求体格式错误", err))
		return
	}
	id, err := s.svc.Alloc.Allocate(r.Context(), &u)
	if err != nil {
		writeError(w, err)
		return
	}
	writeCreated(w, map[string]any{"id": id})
}

func (s *Server) listAllocations(w http.ResponseWriter, r *http.Request) {
	p := pagination(r)
	f := repository.AllocFilter{}
	if v := r.URL.Query().Get("batch_id"); v != "" {
		if id, err := strconv.ParseInt(v, 10, 64); err == nil {
			f.BatchID = &id
		}
	}
	if v := r.URL.Query().Get("task_id"); v != "" {
		if id, err := strconv.ParseInt(v, 10, 64); err == nil {
			f.TaskID = &id
		}
	}
	if v := r.URL.Query().Get("material_id"); v != "" {
		if id, err := strconv.ParseInt(v, 10, 64); err == nil {
			f.MaterialID = &id
		}
	}
	if v := r.URL.Query().Get("type"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			t := int8(n)
			f.Type = &t
		}
	}
	res, err := s.svc.Alloc.List(r.Context(), p, f)
	if err != nil {
		writeError(w, err)
		return
	}
	writePage(w, res)
}

func (s *Server) getAllocation(w http.ResponseWriter, r *http.Request) {
	a, err := s.svc.Alloc.Get(r.Context(), pathID(r))
	if err != nil {
		writeError(w, err)
		return
	}
	writeOK(w, a)
}
