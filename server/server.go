package server

import (
	"net/http"

	"github.com/adrianliechti/llama/config"
	"github.com/adrianliechti/llama/server/oai"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

type Server struct {
	*config.Config
	http.Handler
}

func New(cfg *config.Config) (*Server, error) {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	s := &Server{
		Config:  cfg,
		Handler: r,
	}

	oai, err := oai.New(cfg)

	if err != nil {
		return nil, err
	}

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"*"},

		AllowedMethods: []string{
			http.MethodHead,
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
			http.MethodOptions,
		},

		AllowedHeaders: []string{"*"},

		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Use(s.handleAuth)

	r.Mount("/oai", oai)

	return s, nil
}

func (s *Server) ListenAndServe() error {
	return http.ListenAndServe(s.Address, s)
}

func (s *Server) handleAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		var authorized = len(s.Authorizers) == 0

		for _, a := range s.Authorizers {
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
