package server

import (
	"net/http"

	"chat/pkg/auth"
	"chat/provider"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Server struct {
	router *chi.Mux

	auth     auth.Provider
	provider provider.Provider
}

func New(a auth.Provider, p provider.Provider) *Server {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	s := &Server{
		router: r,

		auth:     a,
		provider: p,
	}

	r.Use(s.handleAuth)

	r.Get("/v1/models", s.handleModels)
	r.Get("/v1/model/{id}", s.handleModel)

	r.Post("/v1/completions", s.handleCompletions)
	r.Post("/v1/chat/completions", s.handleChatCompletions)

	return s
}

func (s *Server) ListenAndServe(addr string) error {
	return http.ListenAndServe(addr, s.router)
}

func (s *Server) handleAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		if err := s.auth.Verify(ctx, r); err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
