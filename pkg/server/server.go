package server

import (
	"net/http"

	"github.com/adrianliechti/llama/config"
	"github.com/adrianliechti/llama/pkg/authorizer"
	"github.com/adrianliechti/llama/pkg/dispatcher"
	"github.com/adrianliechti/llama/pkg/provider"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

type Server struct {
	addr   string
	router *chi.Mux

	provider   provider.Provider
	authorizer authorizer.Provider
}

func New(cfg *config.Config) (*Server, error) {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	provider, err := dispatcher.New(cfg.Providers...)

	if err != nil {
		return nil, err
	}

	s := &Server{
		addr:   cfg.Addr,
		router: r,

		provider:   provider,
		authorizer: cfg.Authorizer,
	}

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"*"},

		AllowedMethods: []string{
			http.MethodHead,
			http.MethodGet,
			http.MethodPost,
			http.MethodOptions,
		},

		AllowedHeaders: []string{"*"},

		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Use(s.handleAuth)

	r.Get("/v1/models", s.handleModels)
	r.Get("/v1/models/{id}", s.handleModel)

	r.Post("/v1/embeddings", s.handleEmbeddings)

	r.Post("/v1/chat/completions", s.handleChatCompletions)

	return s, nil
}

func (s *Server) ListenAndServe() error {
	return http.ListenAndServe(s.addr, s.router)
}

func (s *Server) handleAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		if s.authorizer != nil {
			if err := s.authorizer.Verify(ctx, r); err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
