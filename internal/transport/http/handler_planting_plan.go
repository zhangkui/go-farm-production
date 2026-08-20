package http

import (
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"go-farm-production/internal/domain"
	"go-farm-production/internal/repository"
)

// mountPlantingPlans registers planting-plan routes.
func (s *Server) mountPlantingPlans(r *mux.Router) {
	g := r.PathPrefix("/planting-plans").Subrouter()
	g.Handle("", s.Require(domain.PermPlanView)(http.HandlerFunc(s.listPlantingPlans))).Methods(http.MethodGet)
	g.Handle("", s.Require(domain.PermPlanManage)(s.IdempotencyMiddleware(http.HandlerFunc(s.createPlantingPlan)))).Methods(http.MethodPost)
	g.Handle("/{id:[0-9]+}", s.Require(domain.PermPlanView)(http.HandlerFunc(s.getPlantingPlan))).Methods(http.MethodGet)
	g.Handle("/{id:[0-9]+}", s.Require(domain.PermPlanManage)(http.HandlerFunc(s.updatePlantingPlan))).Methods(http.MethodPut)
	g.Handle("/{id:[0-9]+}", s.Require(domain.PermPlanManage)(http.HandlerFunc(s.deletePlantingPlan))).Methods(http.MethodDelete)
	g.Handle("/{id:[0-9]+}/transition", s.Require(domain.PermPlanManage)(http.HandlerFunc(s.transitionPlantingPlan))).Methods(http.MethodPost)
}

func (s *Server) listPlantingPlans(w http.ResponseWriter, r *http.Request) {
	p := pagination(r)
	f := repository.PlanFilter{}
	if v := r.URL.Query().Get("field_id"); v != "" {
		if id, err := strconv.ParseInt(v, 10, 64); err == nil {
			f.FieldID = &id
		}
	}
	if v := r.URL.Query().Get("season_id"); v != "" {
		if id, err := strconv.ParseInt(v, 10, 64); err == nil {
			f.SeasonID = &id
		}
	}
	if v := r.URL.Query().Get("crop_variety_id"); v != "" {
		if id, err := strconv.ParseInt(v, 10, 64); err == nil {
			f.VarietyID = &id
		}
	}
	if v := r.URL.Query().Get("status"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			st := int8(n)
			f.Status = &st
		}
	}
	res, err := s.svc.Plan.List(r.Context(), p, f)
	if err != nil {
		writeError(w, err)
		return
	}
	writePage(w, res)
}

func (s *Server) createPlantingPlan(w http.ResponseWriter, r *http.Request) {
	var u domain.PlantingPlanUpsert
	if err := parseJSON(r, &u); err != nil {
		writeError(w, domain.Wrap(domain.CodeInvalidJSON, 400, "请求体格式错误", err))
		return
	}
	id, err := s.svc.Plan.Create(r.Context(), &u)
	if err != nil {
		writeError(w, err)
		return
	}
	writeCreated(w, map[string]any{"id": id})
}

func (s *Server) getPlantingPlan(w http.ResponseWriter, r *http.Request) {
	plan, tasks, harvests, err := s.svc.Plan.Get(r.Context(), pathID(r))
	if err != nil {
		writeError(w, err)
		return
	}
	writeOK(w, map[string]any{"plan": plan, "tasks": tasks, "harvests": harvests})
}

func (s *Server) updatePlantingPlan(w http.ResponseWriter, r *http.Request) {
	id := pathID(r)
	var u domain.PlantingPlanUpsert
	if err := parseJSON(r, &u); err != nil {
		writeError(w, domain.Wrap(domain.CodeInvalidJSON, 400, "请求体格式错误", err))
		return
	}
	if err := s.svc.Plan.Update(r.Context(), id, &u); err != nil {
		writeError(w, err)
		return
	}
	writeOK(w, map[string]any{"id": id})
}

func (s *Server) deletePlantingPlan(w http.ResponseWriter, r *http.Request) {
	id := pathID(r)
	if err := s.svc.Plan.Delete(r.Context(), id); err != nil {
		writeError(w, err)
		return
	}
	writeOK(w, map[string]any{"id": id})
}

type transitionRequest struct {
	To int8 `json:"to"`
}

func (s *Server) transitionPlantingPlan(w http.ResponseWriter, r *http.Request) {
	id := pathID(r)
	var req transitionRequest
	if err := parseJSON(r, &req); err != nil {
		writeError(w, domain.Wrap(domain.CodeInvalidJSON, 400, "请求体格式错误", err))
		return
	}
	if err := s.svc.Plan.Transition(r.Context(), id, req.To); err != nil {
		writeError(w, err)
		return
	}
	writeOK(w, map[string]any{"id": id, "status": req.To})
}
