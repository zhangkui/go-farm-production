package http

import (
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"go-farm-production/internal/domain"
)

// mountCostAnalyses registers cost-analysis routes.
func (s *Server) mountCostAnalyses(r *mux.Router) {
	g := r.PathPrefix("/cost-analyses").Subrouter()
	g.Use(s.Require(domain.PermCostView))
	g.Handle("", s.IdempotencyMiddleware(http.HandlerFunc(s.computeCost))).Methods(http.MethodPost)
	g.HandleFunc("", s.listCostAnalyses).Methods(http.MethodGet)
	g.HandleFunc("/{id:[0-9]+}", s.getCostAnalysis).Methods(http.MethodGet)
	g.HandleFunc("/plan/{plan_id:[0-9]+}", s.getCostByPlan).Methods(http.MethodGet)
	g.HandleFunc("/{id:[0-9]+}", s.deleteCostAnalysis).Methods(http.MethodDelete)
}

func (s *Server) computeCost(w http.ResponseWriter, r *http.Request) {
	var u domain.CostAnalysisUpsert
	if err := parseJSON(r, &u); err != nil {
		writeError(w, domain.Wrap(domain.CodeInvalidJSON, 400, "请求体格式错误", err))
		return
	}
	id, err := s.svc.Cost.Compute(r.Context(), &u)
	if err != nil {
		writeError(w, err)
		return
	}
	writeCreated(w, map[string]any{"id": id})
}

func (s *Server) listCostAnalyses(w http.ResponseWriter, r *http.Request) {
	p := pagination(r)
	var planID *int64
	if v := r.URL.Query().Get("planting_plan_id"); v != "" {
		if id, err := strconv.ParseInt(v, 10, 64); err == nil {
			planID = &id
		}
	}
	res, err := s.svc.Cost.List(r.Context(), p, planID)
	if err != nil {
		writeError(w, err)
		return
	}
	writePage(w, res)
}

func (s *Server) getCostAnalysis(w http.ResponseWriter, r *http.Request) {
	c, err := s.svc.Cost.Get(r.Context(), pathID(r))
	if err != nil {
		writeError(w, err)
		return
	}
	writeOK(w, c)
}

func (s *Server) getCostByPlan(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(mux.Vars(r)["plan_id"], 10, 64)
	c, err := s.svc.Cost.GetByPlan(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeOK(w, c)
}

func (s *Server) deleteCostAnalysis(w http.ResponseWriter, r *http.Request) {
	id := pathID(r)
	if err := s.svc.Cost.Delete(r.Context(), id); err != nil {
		writeError(w, err)
		return
	}
	writeOK(w, map[string]any{"id": id})
}
