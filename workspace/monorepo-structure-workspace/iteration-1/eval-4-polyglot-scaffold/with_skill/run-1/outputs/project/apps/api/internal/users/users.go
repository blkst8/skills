// Package users is a feature module: its handlers, request types, and
// validation live together here. New features get their own sibling
// package under internal/ (internal/products, internal/auth, ...).
package users

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/yourname/sideproject/internal/store"
)

// Register mounts the users feature. Callers wrap it in a route group,
// e.g. r.Route("/api/v1", func(r chi.Router) { users.Register(r, db) }).
func Register(r chi.Router, db *store.Store) {
	h := &handler{db: db}
	r.Get("/users", h.list)
	r.Post("/users", h.create)
}

type handler struct {
	db *store.Store
}

func (h *handler) list(w http.ResponseWriter, r *http.Request) {
	users, err := h.db.ListUsers(r.Context())
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, users)
}

type createUserRequest struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

func (h *handler) create(w http.ResponseWriter, r *http.Request) {
	var req createUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if req.Email == "" || req.Name == "" {
		http.Error(w, "email and name are required", http.StatusBadRequest)
		return
	}

	u, err := h.db.CreateUser(r.Context(), req.Email, req.Name)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusCreated, u)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
