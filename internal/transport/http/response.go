package http

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"

	"go-farm-production/internal/domain"
)

// Envelope is the standard JSON response shape (design §9.3).
type Envelope struct {
	Code      int    `json:"code"`
	Message   string `json:"message"`
	Data      any    `json:"data,omitempty"`
	RequestID string `json:"request_id,omitempty"`
}

// writeJSON serialises v and writes it with the given status.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// writeOK returns a 200 with a success envelope.
func writeOK(w http.ResponseWriter, data any) {
	writeJSON(w, http.StatusOK, Envelope{Code: domain.CodeSuccess, Message: "success", Data: data})
}

// writeCreated returns a 201 with the created resource.
func writeCreated(w http.ResponseWriter, data any) {
	writeJSON(w, http.StatusCreated, Envelope{Code: domain.CodeSuccess, Message: "created", Data: data})
}

// writeError maps an error to the standard envelope and HTTP status.
func writeError(w http.ResponseWriter, err error) {
	ae := domain.AsAppError(err)
	writeJSON(w, ae.HTTP, Envelope{Code: ae.Code, Message: ae.Message})
}

// writePage writes a paginated result.
func writePage[T any](w http.ResponseWriter, pr domain.PageResult[T]) {
	writeOK(w, pr)
}

// parseJSON decodes the request body into v. Unknown fields are ignored so the
// frontend can safely re-send a fetched row (which carries id/timestamps) as
// an upsert payload without hitting a strict-mode rejection.
func parseJSON(r *http.Request, v any) error {
	dec := json.NewDecoder(r.Body)
	return dec.Decode(v)
}

// pagination builds a Pagination from query params with safe defaults.
func pagination(r *http.Request) domain.Pagination {
	p := domain.Pagination{
		Page:     atoi(r, "page", domain.DefaultPage),
		PageSize: atoi(r, "page_size", domain.DefaultPageSize),
		OrderBy:  r.URL.Query().Get("order_by"),
		Order:    r.URL.Query().Get("order"),
		Search:   r.URL.Query().Get("search"),
	}
	p.Normalize()
	return p
}

// atoi reads an int query param with a fallback default.
func atoi(r *http.Request, key string, def int) int {
	s := r.URL.Query().Get(key)
	if s == "" {
		return def
	}
	var n int
	for _, c := range s {
		if c < '0' || c > '9' {
			return def
		}
		n = n*10 + int(c-'0')
		if n > 1_000_000 {
			return def
		}
	}
	return n
}

// newReqID generates a request identifier.
func newReqID() string { return uuid.NewString() }
