package server

import (
	"net/http"

	"github.com/adrianliechti/llama/config"
	"github.com/adrianliechti/llama/pkg/auth"
	"github.com/adrianliechti/llama/pkg/llm"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Server struct {
	addr   string
	router *chi.Mux

	auth auth.Provider
	llm  llm.Provider
}

func New(cfg *config.Config) *Server {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	s := &Server{
		addr:   cfg.Addr,
		router: r,

		auth: cfg.Auth,
		llm:  cfg.LLM,
	}

	r.Use(s.handleAuth)

	r.Get("/v1/models", s.handleModels)
	r.Get("/v1/model/{id}", s.handleModel)

	r.Post("/v1/embeddings", s.handleEmbeddings)

	r.Post("/v1/completions", s.handleCompletions)
	r.Post("/v1/chat/completions", s.handleChatCompletions)

	return s
}

func (s *Server) ListenAndServe() error {
	return http.ListenAndServe(s.addr, s.router)
}

func (s *Server) handleAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		if s.auth != nil {
			if err := s.auth.Verify(ctx, r); err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
