package api

import (
	"net/http"

	"github.com/adrianliechti/llama/config"

	"github.com/go-chi/chi/v5"
)

type Server struct {
	*config.Config
	http.Handler
}

func New(cfg *config.Config) (*Server, error) {
	r := chi.NewRouter()

	s := &Server{
		Config:  cfg,
		Handler: r,
	}

	r.Post("/index/{index}", s.handleIndex)

	return s, nil
}
