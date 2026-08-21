package http

import (
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"go-farm-production/internal/domain"
)

// mountProduceInventory registers produce-inventory routes.
func (s *Server) mountProduceInventory(r *mux.Router) {
	g := r.PathPrefix("/produce-inventory").Subrouter()
	g.Handle("", s.Require(domain.PermProduceView)(http.HandlerFunc(s.listProduce))).Methods(http.MethodGet)
	g.Handle("", s.Require(domain.PermProduceManage)(http.HandlerFunc(s.createProduce))).Methods(http.MethodPost)
	g.Handle("/{id:[0-9]+}", s.Require(domain.PermProduceView)(http.HandlerFunc(s.getProduce))).Methods(http.MethodGet)
	g.Handle("/{id:[0-9]+}", s.Require(domain.PermProduceManage)(http.HandlerFunc(s.updateProduce))).Methods(http.MethodPut)
	g.Handle("/{id:[0-9]+}", s.Require(domain.PermProduceManage)(http.HandlerFunc(s.deleteProduce))).Methods(http.MethodDelete)
}

func (s *Server) listProduce(w http.ResponseWriter, r *http.Request) {
	p := pagination(r)
	var status *int8
	var varietyID *int64
	if v := r.URL.Query().Get("status"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			st := int8(n)
			status = &st
		}
	}
	if v := r.URL.Query().Get("crop_variety_id"); v != "" {
		if id, err := strconv.ParseInt(v, 10, 64); err == nil {
			varietyID = &id
		}
	}
	res, err := s.svc.Produce.List(r.Context(), p, status, varietyID)
	if err != nil {
		writeError(w, err)
		return
	}
	writePage(w, res)
}

func (s *Server) createProduce(w http.ResponseWriter, r *http.Request) {
	var u domain.ProduceInventoryUpsert
	if err := parseJSON(r, &u); err != nil {
		writeError(w, domain.Wrap(domain.CodeInvalidJSON, 400, "请求体格式错误", err))
		return
	}
	id, err := s.svc.Produce.Create(r.Context(), &u)
	if err != nil {
		writeError(w, err)
		return
	}
	writeCreated(w, map[string]any{"id": id})
}

func (s *Server) getProduce(w http.ResponseWriter, r *http.Request) {
	p, err := s.svc.Produce.Get(r.Context(), pathID(r))
	if err != nil {
		writeError(w, err)
		return
	}
	writeOK(w, p)
}

func (s *Server) updateProduce(w http.ResponseWriter, r *http.Request) {
	id := pathID(r)
	var u domain.ProduceInventoryUpsert
	if err := parseJSON(r, &u); err != nil {
		writeError(w, domain.Wrap(domain.CodeInvalidJSON, 400, "请求体格式错误", err))
		return
	}
	if err := s.svc.Produce.Update(r.Context(), id, &u); err != nil {
		writeError(w, err)
		return
	}
	writeOK(w, map[string]any{"id": id})
}

func (s *Server) deleteProduce(w http.ResponseWriter, r *http.Request) {
	id := pathID(r)
	if err := s.svc.Produce.Delete(r.Context(), id); err != nil {
		writeError(w, err)
		return
	}
	writeOK(w, map[string]any{"id": id})
}
