package server

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/adrianliechti/llama/pkg/provider"

	"github.com/go-chi/chi/v5"
	"github.com/sashabaranov/go-openai"
)

func (s *Server) handleModels(w http.ResponseWriter, r *http.Request) {
	models, err := s.provider.Models(r.Context())

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	result := openai.ModelsList{
		Models: convertModels(models),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (s *Server) handleModel(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	models, err := s.provider.Models(r.Context())

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	for _, m := range models {
		if !strings.EqualFold(id, m.ID) {
			continue
		}

		result := convertModel(m)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
		return
	}

	w.WriteHeader(http.StatusNotFound)
}

func convertModels(s []provider.Model) []openai.Model {
	var result []openai.Model

	for _, m := range s {
		result = append(result, convertModel(m))
	}

	return result
}

func convertModel(m provider.Model) openai.Model {
	return openai.Model{
		ID: m.ID,

		Object:    "model",
		CreatedAt: time.Now().Unix(),
	}
}
