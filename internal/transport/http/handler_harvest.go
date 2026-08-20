package http

import (
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"go-farm-production/internal/domain"
	"go-farm-production/internal/repository"
)

// mountHarvests registers harvest routes.
func (s *Server) mountHarvests(r *mux.Router) {
	g := r.PathPrefix("/harvests").Subrouter()
	g.Handle("", s.Require(domain.PermHarvestView)(http.HandlerFunc(s.listHarvests))).Methods(http.MethodGet)
	g.Handle("", s.Require(domain.PermHarvestManage)(s.IdempotencyMiddleware(http.HandlerFunc(s.createHarvest)))).Methods(http.MethodPost)
	g.Handle("/{id:[0-9]+}", s.Require(domain.PermHarvestView)(http.HandlerFunc(s.getHarvest))).Methods(http.MethodGet)
	g.Handle("/{id:[0-9]+}", s.Require(domain.PermHarvestManage)(http.HandlerFunc(s.updateHarvest))).Methods(http.MethodPut)
	g.Handle("/{id:[0-9]+}", s.Require(domain.PermHarvestManage)(http.HandlerFunc(s.deleteHarvest))).Methods(http.MethodDelete)
	g.Handle("/{id:[0-9]+}/approve", s.Require(domain.PermHarvestApprove)(http.HandlerFunc(s.approveHarvest))).Methods(http.MethodPost)
}

func (s *Server) listHarvests(w http.ResponseWriter, r *http.Request) {
	p := pagination(r)
	f := repository.HarvestFilter{}
	if v := r.URL.Query().Get("planting_plan_id"); v != "" {
		if id, err := strconv.ParseInt(v, 10, 64); err == nil {
			f.PlanID = &id
		}
	}
	if v := r.URL.Query().Get("approved"); v != "" {
		b := v == "true" || v == "1"
		f.Approved = &b
	}
	res, err := s.svc.Harvest.List(r.Context(), p, f)
	if err != nil {
		writeError(w, err)
		return
	}
	writePage(w, res)
}

func (s *Server) createHarvest(w http.ResponseWriter, r *http.Request) {
	var u domain.HarvestUpsert
	if err := parseJSON(r, &u); err != nil {
		writeError(w, domain.Wrap(domain.CodeInvalidJSON, 400, "请求体格式错误", err))
		return
	}
	id, err := s.svc.Harvest.Create(r.Context(), &u)
	if err != nil {
		writeError(w, err)
		return
	}
	writeCreated(w, map[string]any{"id": id})
}

func (s *Server) getHarvest(w http.ResponseWriter, r *http.Request) {
	h, details, err := s.svc.Harvest.Get(r.Context(), pathID(r))
	if err != nil {
		writeError(w, err)
		return
	}
	writeOK(w, map[string]any{"harvest": h, "details": details})
}

func (s *Server) updateHarvest(w http.ResponseWriter, r *http.Request) {
	id := pathID(r)
	var u domain.HarvestUpsert
	if err := parseJSON(r, &u); err != nil {
		writeError(w, domain.Wrap(domain.CodeInvalidJSON, 400, "请求体格式错误", err))
		return
	}
	if err := s.svc.Harvest.Update(r.Context(), id, &u); err != nil {
		writeError(w, err)
		return
	}
	writeOK(w, map[string]any{"id": id})
}

func (s *Server) deleteHarvest(w http.ResponseWriter, r *http.Request) {
	id := pathID(r)
	if err := s.svc.Harvest.Delete(r.Context(), id); err != nil {
		writeError(w, err)
		return
	}
	writeOK(w, map[string]any{"id": id})
}

func (s *Server) approveHarvest(w http.ResponseWriter, r *http.Request) {
	id := pathID(r)
	if err := s.svc.Harvest.Approve(r.Context(), id); err != nil {
		writeError(w, err)
		return
	}
	writeOK(w, map[string]any{"id": id, "approved": true})
}
