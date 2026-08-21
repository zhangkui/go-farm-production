package http

import (
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"go-farm-production/internal/domain"
)

// mountSeasons registers season routes.
func (s *Server) mountSeasons(r *mux.Router) {
	g := r.PathPrefix("/seasons").Subrouter()
	g.Handle("", s.Require(domain.PermSeasonView)(http.HandlerFunc(s.listSeasons))).Methods(http.MethodGet)
	g.Handle("", s.Require(domain.PermSeasonManage)(http.HandlerFunc(s.createSeason))).Methods(http.MethodPost)
	g.Handle("/{id:[0-9]+}", s.Require(domain.PermSeasonView)(http.HandlerFunc(s.getSeason))).Methods(http.MethodGet)
	g.Handle("/{id:[0-9]+}", s.Require(domain.PermSeasonManage)(http.HandlerFunc(s.updateSeason))).Methods(http.MethodPut)
	g.Handle("/{id:[0-9]+}", s.Require(domain.PermSeasonManage)(http.HandlerFunc(s.deleteSeason))).Methods(http.MethodDelete)
}

func (s *Server) listSeasons(w http.ResponseWriter, r *http.Request) {
	p := pagination(r)
	var status *int8
	if v := r.URL.Query().Get("status"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			s := int8(n)
			status = &s
		}
	}
	res, err := s.svc.Season.List(r.Context(), p, status)
	if err != nil {
		writeError(w, err)
		return
	}
	writePage(w, res)
}

func (s *Server) createSeason(w http.ResponseWriter, r *http.Request) {
	var u domain.SeasonUpsert
	if err := parseJSON(r, &u); err != nil {
		writeError(w, domain.Wrap(domain.CodeInvalidJSON, 400, "请求体格式错误", err))
		return
	}
	id, err := s.svc.Season.Create(r.Context(), &u)
	if err != nil {
		writeError(w, err)
		return
	}
	writeCreated(w, map[string]any{"id": id})
}

func (s *Server) getSeason(w http.ResponseWriter, r *http.Request) {
	v, err := s.svc.Season.Get(r.Context(), pathID(r))
	if err != nil {
		writeError(w, err)
		return
	}
	writeOK(w, v)
}

func (s *Server) updateSeason(w http.ResponseWriter, r *http.Request) {
	id := pathID(r)
	var u domain.SeasonUpsert
	if err := parseJSON(r, &u); err != nil {
		writeError(w, domain.Wrap(domain.CodeInvalidJSON, 400, "请求体格式错误", err))
		return
	}
	if err := s.svc.Season.Update(r.Context(), id, &u); err != nil {
		writeError(w, err)
		return
	}
	writeOK(w, map[string]any{"id": id})
}

func (s *Server) deleteSeason(w http.ResponseWriter, r *http.Request) {
	id := pathID(r)
	if err := s.svc.Season.Delete(r.Context(), id); err != nil {
		writeError(w, err)
		return
	}
	writeOK(w, map[string]any{"id": id})
}
