package server

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/sashabaranov/go-openai"
)

func (s *Server) handleModels(w http.ResponseWriter, r *http.Request) {
	data, err := s.provider.Models(r.Context())

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	result := openai.ModelsList{}

	for _, m := range data {
		result.Models = append(result.Models, openai.Model{
			ID: m.ID,

			Object:    "model",
			CreatedAt: time.Now().Unix(),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (s *Server) handleModel(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	data, err := s.provider.Models(r.Context())

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	for _, m := range data {
		if !strings.EqualFold(id, m.ID) {
			continue
		}

		result := openai.Model{
			ID: m.ID,

			Object:    "model",
			CreatedAt: time.Now().Unix(),
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
		return
	}

	w.WriteHeader(http.StatusNotFound)
}
