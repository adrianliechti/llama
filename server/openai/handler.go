package openai

import (
	"encoding/json"
	"net/http"

	"github.com/adrianliechti/llama/config"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	*config.Config
	http.Handler
}

func New(cfg *config.Config) (*Handler, error) {
	mux := chi.NewMux()

	h := &Handler{
		Config:  cfg,
		Handler: mux,
	}

	h.Attach(mux)
	return h, nil
}

func (h *Handler) Attach(r chi.Router) {
	r.Get("/models", h.handleModels)
	r.Get("/models/{id}", h.handleModel)

	r.Post("/embeddings", h.handleEmbeddings)

	r.Post("/chat/completions", h.handleChatCompletion)

	r.Post("/audio/speech", h.handleAudioSpeech)
	r.Post("/audio/transcriptions", h.handleAudioTranscription)

	r.Post("/images/generations", h.handleImageGeneration)
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
