package oai

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

	r.Get("/v1/models", s.handleModels)
	r.Get("/v1/models/{id}", s.handleModel)

	r.Post("/v1/embeddings", s.handleEmbeddings)

	r.Post("/v1/chat/completions", s.handleChatCompletions)
	r.Post("/v1/audio/transcriptions", s.handleAudioTranscriptions)

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

	errorType := "invalid_request_error"

	if code >= 500 {
		errorType = "internal_server_error"
	}

	resp := ErrorResponse{
		Error: Error{
			Type:    errorType,
			Message: err.Error(),
		},
	}

	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)

	enc.Encode(resp)
}
