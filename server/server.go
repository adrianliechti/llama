package server

import (
	"net/http"

	"github.com/adrianliechti/llama/config"
	"github.com/adrianliechti/llama/server/api"
	"github.com/adrianliechti/llama/server/openai"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

type Server struct {
	*config.Config
	http.Handler

	api    *api.Handler
	openai *openai.Handler
}

func New(cfg *config.Config) (*Server, error) {
	api, err := api.New(cfg)

	if err != nil {
		return nil, err
	}

	openai, err := openai.New(cfg)

	if err != nil {
		return nil, err
	}

	mux := chi.NewMux()

	s := &Server{
		Config:  cfg,
		Handler: mux,

		api:    api,
		openai: openai,
	}

	mux.Use(middleware.Logger)
	mux.Use(middleware.Recoverer)

	mux.Use(cors.Handler(cors.Options{
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

		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,

		MaxAge: 300,
	}))

	mux.Use(otelhttp.NewMiddleware("http"))

	mux.Use(s.handleAuth)

	mux.Route("/v1", func(r chi.Router) {
		s.api.Attach(r)
		s.openai.Attach(r)
	})

	mux.Route("/api", func(r chi.Router) {
		s.api.Attach(r)
	})

	mux.Route("/oai/v1", func(r chi.Router) {
		s.openai.Attach(r)
	})

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
