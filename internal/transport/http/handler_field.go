package http

import (
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"go-farm-production/internal/domain"
)

// mountFields registers field routes.
func (s *Server) mountFields(r *mux.Router) {
	g := r.PathPrefix("/fields").Subrouter()
	g.Handle("", s.Require(domain.PermFieldView)(http.HandlerFunc(s.listFields))).Methods(http.MethodGet)
	g.Handle("", s.Require(domain.PermFieldManage)(http.HandlerFunc(s.createField))).Methods(http.MethodPost)
	g.Handle("/{id:[0-9]+}", s.Require(domain.PermFieldView)(http.HandlerFunc(s.getField))).Methods(http.MethodGet)
	g.Handle("/{id:[0-9]+}", s.Require(domain.PermFieldManage)(http.HandlerFunc(s.updateField))).Methods(http.MethodPut)
	g.Handle("/{id:[0-9]+}", s.Require(domain.PermFieldManage)(http.HandlerFunc(s.deleteField))).Methods(http.MethodDelete)
}

func (s *Server) listFields(w http.ResponseWriter, r *http.Request) {
	p := pagination(r)
	var farmID *int64
	if v := r.URL.Query().Get("farm_id"); v != "" {
		if id, err := strconv.ParseInt(v, 10, 64); err == nil {
			farmID = &id
		}
	}
	zone := r.URL.Query().Get("irrigation_zone")
	res, err := s.svc.Field.List(r.Context(), p, farmID, zone)
	if err != nil {
		writeError(w, err)
		return
	}
	writePage(w, res)
}

func (s *Server) createField(w http.ResponseWriter, r *http.Request) {
	var u domain.FieldUpsert
	if err := parseJSON(r, &u); err != nil {
		writeError(w, domain.Wrap(domain.CodeInvalidJSON, 400, "请求体格式错误", err))
		return
	}
	id, err := s.svc.Field.Create(r.Context(), &u)
	if err != nil {
		writeError(w, err)
		return
	}
	writeCreated(w, map[string]any{"id": id})
}

func (s *Server) getField(w http.ResponseWriter, r *http.Request) {
	id := pathID(r)
	f, err := s.svc.Field.Get(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeOK(w, f)
}

func (s *Server) updateField(w http.ResponseWriter, r *http.Request) {
	id := pathID(r)
	var u domain.FieldUpsert
	if err := parseJSON(r, &u); err != nil {
		writeError(w, domain.Wrap(domain.CodeInvalidJSON, 400, "请求体格式错误", err))
		return
	}
	if err := s.svc.Field.Update(r.Context(), id, &u); err != nil {
		writeError(w, err)
		return
	}
	writeOK(w, map[string]any{"id": id})
}

func (s *Server) deleteField(w http.ResponseWriter, r *http.Request) {
	id := pathID(r)
	if err := s.svc.Field.Delete(r.Context(), id); err != nil {
		writeError(w, err)
		return
	}
	writeOK(w, map[string]any{"id": id})
}
