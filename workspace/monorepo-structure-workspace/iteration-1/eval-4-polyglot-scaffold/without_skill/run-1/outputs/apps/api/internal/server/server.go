package server

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/blkst8/sideproject/apps/api/internal/apitypes"
	"github.com/blkst8/sideproject/apps/api/internal/store"
)

type Server struct {
	store store.Store
}

func New(st store.Store) http.Handler {
	s := &Server{store: st}

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Route("/api", func(r chi.Router) {
		r.Get("/healthz", s.handleHealth)
		r.Get("/items", s.handleListItems)
	})

	return r
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	// The DB flag is static-ok for now; ping the pool from the handler once
	// real dependencies are wired in.
	writeJSON(w, http.StatusOK, apitypes.HealthStatus{Status: "ok", DB: true})
}

func (s *Server) handleListItems(w http.ResponseWriter, r *http.Request) {
	items, err := s.store.ListItems(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal"})
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}
