package http

import (
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"go-farm-production/internal/domain"
	"go-farm-production/internal/repository"
)

// mountTasks registers farm-task routes.
func (s *Server) mountTasks(r *mux.Router) {
	g := r.PathPrefix("/tasks").Subrouter()
	g.Handle("", s.Require(domain.PermTaskView)(http.HandlerFunc(s.listTasks))).Methods(http.MethodGet)
	g.Handle("", s.Require(domain.PermTaskManage)(s.IdempotencyMiddleware(http.HandlerFunc(s.createTask)))).Methods(http.MethodPost)
	g.Handle("/{id:[0-9]+}", s.Require(domain.PermTaskView)(http.HandlerFunc(s.getTask))).Methods(http.MethodGet)
	g.Handle("/{id:[0-9]+}", s.Require(domain.PermTaskManage)(http.HandlerFunc(s.updateTask))).Methods(http.MethodPut)
	g.Handle("/{id:[0-9]+}", s.Require(domain.PermTaskManage)(http.HandlerFunc(s.deleteTask))).Methods(http.MethodDelete)
	g.Handle("/{id:[0-9]+}/transition", s.Require(domain.PermTaskExecute)(http.HandlerFunc(s.transitionTask))).Methods(http.MethodPost)
}

func (s *Server) listTasks(w http.ResponseWriter, r *http.Request) {
	p := pagination(r)
	f := repository.TaskFilter{}
	if v := r.URL.Query().Get("planting_plan_id"); v != "" {
		if id, err := strconv.ParseInt(v, 10, 64); err == nil {
			f.PlanID = &id
		}
	}
	if v := r.URL.Query().Get("status"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			st := int8(n)
			f.Status = &st
		}
	}
	if v := r.URL.Query().Get("task_type"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			tt := int8(n)
			f.TaskType = &tt
		}
	}
	if v := r.URL.Query().Get("assignee_id"); v != "" {
		if id, err := strconv.ParseInt(v, 10, 64); err == nil {
			f.AssigneeID = &id
		}
	}
	res, err := s.svc.Task.List(r.Context(), p, f)
	if err != nil {
		writeError(w, err)
		return
	}
	writePage(w, res)
}

func (s *Server) createTask(w http.ResponseWriter, r *http.Request) {
	var u domain.FarmTaskUpsert
	if err := parseJSON(r, &u); err != nil {
		writeError(w, domain.Wrap(domain.CodeInvalidJSON, 400, "请求体格式错误", err))
		return
	}
	id, err := s.svc.Task.Create(r.Context(), &u)
	if err != nil {
		writeError(w, err)
		return
	}
	writeCreated(w, map[string]any{"id": id})
}

func (s *Server) getTask(w http.ResponseWriter, r *http.Request) {
	t, allocs, err := s.svc.Task.Get(r.Context(), pathID(r))
	if err != nil {
		writeError(w, err)
		return
	}
	writeOK(w, map[string]any{"task": t, "allocations": allocs})
}

func (s *Server) updateTask(w http.ResponseWriter, r *http.Request) {
	id := pathID(r)
	var u domain.FarmTaskUpsert
	if err := parseJSON(r, &u); err != nil {
		writeError(w, domain.Wrap(domain.CodeInvalidJSON, 400, "请求体格式错误", err))
		return
	}
	if err := s.svc.Task.Update(r.Context(), id, &u); err != nil {
		writeError(w, err)
		return
	}
	writeOK(w, map[string]any{"id": id})
}

func (s *Server) deleteTask(w http.ResponseWriter, r *http.Request) {
	id := pathID(r)
	if err := s.svc.Task.Delete(r.Context(), id); err != nil {
		writeError(w, err)
		return
	}
	writeOK(w, map[string]any{"id": id})
}

func (s *Server) transitionTask(w http.ResponseWriter, r *http.Request) {
	id := pathID(r)
	var req transitionRequest
	if err := parseJSON(r, &req); err != nil {
		writeError(w, domain.Wrap(domain.CodeInvalidJSON, 400, "请求体格式错误", err))
		return
	}
	if err := s.svc.Task.Transition(r.Context(), id, req.To); err != nil {
		writeError(w, err)
		return
	}
	writeOK(w, map[string]any{"id": id, "status": req.To})
}
