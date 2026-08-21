package http

import (
	"net/http"

	"github.com/gorilla/mux"

	"go-farm-production/internal/domain"
)

// mountCropVarieties registers crop-variety routes.
func (s *Server) mountCropVarieties(r *mux.Router) {
	g := r.PathPrefix("/crop-varieties").Subrouter()
	g.Handle("", s.Require(domain.PermVarietyView)(http.HandlerFunc(s.listCropVarieties))).Methods(http.MethodGet)
	g.Handle("", s.Require(domain.PermVarietyManage)(http.HandlerFunc(s.createCropVariety))).Methods(http.MethodPost)
	g.Handle("/{id:[0-9]+}", s.Require(domain.PermVarietyView)(http.HandlerFunc(s.getCropVariety))).Methods(http.MethodGet)
	g.Handle("/{id:[0-9]+}", s.Require(domain.PermVarietyManage)(http.HandlerFunc(s.updateCropVariety))).Methods(http.MethodPut)
	g.Handle("/{id:[0-9]+}", s.Require(domain.PermVarietyManage)(http.HandlerFunc(s.deleteCropVariety))).Methods(http.MethodDelete)
}

func (s *Server) listCropVarieties(w http.ResponseWriter, r *http.Request) {
	p := pagination(r)
	category := r.URL.Query().Get("category")
	res, err := s.svc.Variety.List(r.Context(), p, category)
	if err != nil {
		writeError(w, err)
		return
	}
	writePage(w, res)
}

func (s *Server) createCropVariety(w http.ResponseWriter, r *http.Request) {
	var u domain.CropVarietyUpsert
	if err := parseJSON(r, &u); err != nil {
		writeError(w, domain.Wrap(domain.CodeInvalidJSON, 400, "请求体格式错误", err))
		return
	}
	id, err := s.svc.Variety.Create(r.Context(), &u)
	if err != nil {
		writeError(w, err)
		return
	}
	writeCreated(w, map[string]any{"id": id})
}

func (s *Server) getCropVariety(w http.ResponseWriter, r *http.Request) {
	v, err := s.svc.Variety.Get(r.Context(), pathID(r))
	if err != nil {
		writeError(w, err)
		return
	}
	writeOK(w, v)
}

func (s *Server) updateCropVariety(w http.ResponseWriter, r *http.Request) {
	id := pathID(r)
	var u domain.CropVarietyUpsert
	if err := parseJSON(r, &u); err != nil {
		writeError(w, domain.Wrap(domain.CodeInvalidJSON, 400, "请求体格式错误", err))
		return
	}
	if err := s.svc.Variety.Update(r.Context(), id, &u); err != nil {
		writeError(w, err)
		return
	}
	writeOK(w, map[string]any{"id": id})
}

func (s *Server) deleteCropVariety(w http.ResponseWriter, r *http.Request) {
	id := pathID(r)
	if err := s.svc.Variety.Delete(r.Context(), id); err != nil {
		writeError(w, err)
		return
	}
	writeOK(w, map[string]any{"id": id})
}
