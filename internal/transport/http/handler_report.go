package http

import (
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"go-farm-production/internal/domain"
)

// mountReports registers read-only reporting routes.
func (s *Server) mountReports(r *mux.Router) {
	g := r.PathPrefix("/reports").Subrouter()
	g.Use(s.Require(domain.PermCostView))
	g.HandleFunc("/cost-summary", s.reportCostSummary).Methods(http.MethodGet)
	g.HandleFunc("/field-productivity", s.reportFieldProductivity).Methods(http.MethodGet)
}

// reportCostSummary returns cost rollups across plans, optionally filtered by
// season and/or crop variety.
func (s *Server) reportCostSummary(w http.ResponseWriter, r *http.Request) {
	var seasonID, varietyID *int64
	if v := r.URL.Query().Get("season_id"); v != "" {
		if id, err := strconv.ParseInt(v, 10, 64); err == nil {
			seasonID = &id
		}
	}
	if v := r.URL.Query().Get("crop_variety_id"); v != "" {
		if id, err := strconv.ParseInt(v, 10, 64); err == nil {
			varietyID = &id
		}
	}
	items, err := s.svc.Cost.Summarise(r.Context(), seasonID, varietyID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeOK(w, map[string]any{"items": items})
}

// reportFieldProductivity is an alias for the same rollup sorted by unit cost;
// the front-end renders it as a field productivity leaderboard.
func (s *Server) reportFieldProductivity(w http.ResponseWriter, r *http.Request) {
	var varietyID *int64
	if v := r.URL.Query().Get("crop_variety_id"); v != "" {
		if id, err := strconv.ParseInt(v, 10, 64); err == nil {
			varietyID = &id
		}
	}
	items, err := s.svc.Cost.Summarise(r.Context(), nil, varietyID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeOK(w, map[string]any{"items": items})
}

// _ keeps mux referenced when the router helpers are external.
var _ = mux.NewRouter
