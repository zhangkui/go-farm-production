package http

import (
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"go-farm-production/internal/domain"
)

// mountMaterials registers material routes.
func (s *Server) mountMaterials(r *mux.Router) {
	g := r.PathPrefix("/materials").Subrouter()
	g.Handle("", s.Require(domain.PermInventoryView)(http.HandlerFunc(s.listMaterials))).Methods(http.MethodGet)
	g.Handle("", s.Require(domain.PermMaterialManage)(http.HandlerFunc(s.createMaterial))).Methods(http.MethodPost)
	g.Handle("/{id:[0-9]+}", s.Require(domain.PermInventoryView)(http.HandlerFunc(s.getMaterial))).Methods(http.MethodGet)
	g.Handle("/{id:[0-9]+}", s.Require(domain.PermMaterialManage)(http.HandlerFunc(s.updateMaterial))).Methods(http.MethodPut)
	g.Handle("/{id:[0-9]+}", s.Require(domain.PermMaterialManage)(http.HandlerFunc(s.deleteMaterial))).Methods(http.MethodDelete)
}

func (s *Server) listMaterials(w http.ResponseWriter, r *http.Request) {
	p := pagination(r)
	var cat *int8
	if v := r.URL.Query().Get("category"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			c := int8(n)
			cat = &c
		}
	}
	res, err := s.svc.Material.List(r.Context(), p, cat)
	if err != nil {
		writeError(w, err)
		return
	}
	writePage(w, res)
}

func (s *Server) createMaterial(w http.ResponseWriter, r *http.Request) {
	var u domain.MaterialUpsert
	if err := parseJSON(r, &u); err != nil {
		writeError(w, domain.Wrap(domain.CodeInvalidJSON, 400, "请求体格式错误", err))
		return
	}
	id, err := s.svc.Material.Create(r.Context(), &u)
	if err != nil {
		writeError(w, err)
		return
	}
	writeCreated(w, map[string]any{"id": id})
}

func (s *Server) getMaterial(w http.ResponseWriter, r *http.Request) {
	m, err := s.svc.Material.Get(r.Context(), pathID(r))
	if err != nil {
		writeError(w, err)
		return
	}
	writeOK(w, m)
}

func (s *Server) updateMaterial(w http.ResponseWriter, r *http.Request) {
	id := pathID(r)
	var u domain.MaterialUpsert
	if err := parseJSON(r, &u); err != nil {
		writeError(w, domain.Wrap(domain.CodeInvalidJSON, 400, "请求体格式错误", err))
		return
	}
	if err := s.svc.Material.Update(r.Context(), id, &u); err != nil {
		writeError(w, err)
		return
	}
	writeOK(w, map[string]any{"id": id})
}

func (s *Server) deleteMaterial(w http.ResponseWriter, r *http.Request) {
	id := pathID(r)
	if err := s.svc.Material.Delete(r.Context(), id); err != nil {
		writeError(w, err)
		return
	}
	writeOK(w, map[string]any{"id": id})
}
