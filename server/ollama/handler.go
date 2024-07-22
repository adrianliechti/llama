package ollama

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
	r.Head("/", h.handleHeartbeat)
	r.Get("/", h.handleIndex)

	r.Get("/api/tags", h.handleTags)

	r.Post("/api/chat", h.handleChat)
	r.Post("/api/embeddings", h.handleEmbeddings)
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
