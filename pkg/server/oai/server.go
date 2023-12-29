package oai

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

	r.Get("/v1/models", s.handleModels)
	r.Get("/v1/models/{id}", s.handleModel)

	r.Post("/v1/embeddings", s.handleEmbeddings)

	r.Post("/v1/chat/completions", s.handleChatCompletions)

	return s, nil
}
