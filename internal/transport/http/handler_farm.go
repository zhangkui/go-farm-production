package http

import (
	"net/http"

	"github.com/gorilla/mux"

	"go-farm-production/internal/domain"
)

// mountFarms registers farm routes.
func (s *Server) mountFarms(r *mux.Router) {
	g := r.PathPrefix("/farms").Subrouter()
	// List/Get require farm:view; write ops require farm:manage.
	g.Handle("", s.Require(domain.PermFarmView)(http.HandlerFunc(s.listFarms))).Methods(http.MethodGet)
	g.Handle("", s.Require(domain.PermFarmManage)(http.HandlerFunc(s.createFarm))).Methods(http.MethodPost)
	g.Handle("/{id:[0-9]+}", s.Require(domain.PermFarmView)(http.HandlerFunc(s.getFarm))).Methods(http.MethodGet)
	g.Handle("/{id:[0-9]+}", s.Require(domain.PermFarmManage)(http.HandlerFunc(s.updateFarm))).Methods(http.MethodPut)
	g.Handle("/{id:[0-9]+}", s.Require(domain.PermFarmManage)(http.HandlerFunc(s.deleteFarm))).Methods(http.MethodDelete)
}

func (s *Server) listFarms(w http.ResponseWriter, r *http.Request) {
	p := pagination(r)
	res, err := s.svc.Farm.List(r.Context(), p, p.Search)
	if err != nil {
		writeError(w, err)
		return
	}
	writePage(w, res)
}

func (s *Server) createFarm(w http.ResponseWriter, r *http.Request) {
	var u domain.FarmUpsert
	if err := parseJSON(r, &u); err != nil {
		writeError(w, domain.Wrap(domain.CodeInvalidJSON, 400, "请求体格式错误", err))
		return
	}
	id, err := s.svc.Farm.Create(r.Context(), &u)
	if err != nil {
		writeError(w, err)
		return
	}
	writeCreated(w, map[string]any{"id": id})
}

func (s *Server) getFarm(w http.ResponseWriter, r *http.Request) {
	id := pathID(r)
	f, err := s.svc.Farm.Get(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeOK(w, f)
}

func (s *Server) updateFarm(w http.ResponseWriter, r *http.Request) {
	id := pathID(r)
	var u domain.FarmUpsert
	if err := parseJSON(r, &u); err != nil {
		writeError(w, domain.Wrap(domain.CodeInvalidJSON, 400, "请求体格式错误", err))
		return
	}
	if err := s.svc.Farm.Update(r.Context(), id, &u); err != nil {
		writeError(w, err)
		return
	}
	writeOK(w, map[string]any{"id": id})
}

func (s *Server) deleteFarm(w http.ResponseWriter, r *http.Request) {
	id := pathID(r)
	if err := s.svc.Farm.Delete(r.Context(), id); err != nil {
		writeError(w, err)
		return
	}
	writeOK(w, map[string]any{"id": id})
}
