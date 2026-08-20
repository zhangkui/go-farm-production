package http

import (
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"go-farm-production/internal/domain"
)

// mountBatches registers input-batch routes and the inventory-warning endpoint.
func (s *Server) mountBatches(r *mux.Router) {
	g := r.PathPrefix("/batches").Subrouter()
	g.Handle("", s.Require(domain.PermInventoryView)(http.HandlerFunc(s.listBatches))).Methods(http.MethodGet)
	g.Handle("", s.Require(domain.PermBatchManage)(s.IdempotencyMiddleware(http.HandlerFunc(s.createBatch)))).Methods(http.MethodPost)
	g.Handle("/{id:[0-9]+}", s.Require(domain.PermInventoryView)(http.HandlerFunc(s.getBatch))).Methods(http.MethodGet)
	g.Handle("/{id:[0-9]+}", s.Require(domain.PermBatchManage)(http.HandlerFunc(s.updateBatch))).Methods(http.MethodPut)
	g.Handle("/{id:[0-9]+}", s.Require(domain.PermBatchManage)(http.HandlerFunc(s.deleteBatch))).Methods(http.MethodDelete)

	inv := r.PathPrefix("/inventory").Subrouter()
	inv.Use(s.Require(domain.PermInventoryView))
	inv.HandleFunc("/warnings", s.inventoryWarnings).Methods(http.MethodGet)
}

func (s *Server) listBatches(w http.ResponseWriter, r *http.Request) {
	p := pagination(r)
	var matID, status *int64
	if v := r.URL.Query().Get("material_id"); v != "" {
		if id, err := strconv.ParseInt(v, 10, 64); err == nil {
			matID = &id
		}
	}
	if v := r.URL.Query().Get("status"); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			status = &n
		}
	}
	_ = matID
	// reuse service signature: materialID *int64, status *int8
	var status8 *int8
	if status != nil {
		s := int8(*status)
		status8 = &s
	}
	res, err := s.svc.Batch.List(r.Context(), p, matID, status8)
	if err != nil {
		writeError(w, err)
		return
	}
	writePage(w, res)
}

func (s *Server) createBatch(w http.ResponseWriter, r *http.Request) {
	var u domain.InputBatchUpsert
	if err := parseJSON(r, &u); err != nil {
		writeError(w, domain.Wrap(domain.CodeInvalidJSON, 400, "请求体格式错误", err))
		return
	}
	id, err := s.svc.Batch.Create(r.Context(), &u)
	if err != nil {
		writeError(w, err)
		return
	}
	writeCreated(w, map[string]any{"id": id})
}

func (s *Server) getBatch(w http.ResponseWriter, r *http.Request) {
	b, allocs, err := s.svc.Batch.Get(r.Context(), pathID(r))
	if err != nil {
		writeError(w, err)
		return
	}
	writeOK(w, map[string]any{"batch": b, "allocations": allocs})
}

func (s *Server) updateBatch(w http.ResponseWriter, r *http.Request) {
	id := pathID(r)
	var u domain.InputBatchUpsert
	if err := parseJSON(r, &u); err != nil {
		writeError(w, domain.Wrap(domain.CodeInvalidJSON, 400, "请求体格式错误", err))
		return
	}
	if err := s.svc.Batch.Update(r.Context(), id, &u); err != nil {
		writeError(w, err)
		return
	}
	writeOK(w, map[string]any{"id": id})
}

func (s *Server) deleteBatch(w http.ResponseWriter, r *http.Request) {
	id := pathID(r)
	if err := s.svc.Batch.Delete(r.Context(), id); err != nil {
		writeError(w, err)
		return
	}
	writeOK(w, map[string]any{"id": id})
}

func (s *Server) inventoryWarnings(w http.ResponseWriter, r *http.Request) {
	days := 30
	if v := r.URL.Query().Get("within_days"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			days = n
		}
	}
	items, err := s.svc.Batch.ExpiryWarnings(r.Context(), days)
	if err != nil {
		writeError(w, err)
		return
	}
	writeOK(w, map[string]any{"items": items})
}
