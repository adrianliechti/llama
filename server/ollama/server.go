package ollama

import (
	"encoding/json"
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

	r.Head("/", s.handleHeartbeat)
	r.Get("/", s.handleIndex)

	r.Get("/api/tags", s.handleTags)

	r.Post("/api/chat", s.handleChat)
	r.Post("/api/embeddings", s.handleEmbeddings)

	return s, nil
}

func writeJson(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")

	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)

	enc.Encode(v)
}

func writeError(w http.ResponseWriter, code int, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	resp := StatusError{
		StatusCode: code,

		ErrorMessage: err.Error(),
	}

	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)

	enc.Encode(resp)
}
