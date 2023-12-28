package server

import (
	"net/http"

	"github.com/adrianliechti/llama/config"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

type Server struct {
	*config.Config

	router *chi.Mux
}

func New(cfg *config.Config) (*Server, error) {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	s := &Server{
		Config: cfg,

		router: r,
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
	return http.ListenAndServe(s.Config.Address, s.router)
}

func (s *Server) handleAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		var authorized = len(s.Authorizer) == 0

		for _, a := range s.Authorizer {
			if err := a.Verify(ctx, r); err == nil {
				authorized = true
				break
			}
		}

		if !authorized {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
