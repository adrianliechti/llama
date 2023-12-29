package server

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/sashabaranov/go-openai"
)

func (s *Server) handleModels(w http.ResponseWriter, r *http.Request) {
	models := s.Models()

	result := openai.ModelsList{}

	for _, m := range models {
		result.Models = append(result.Models, openai.Model{
			Object: "model",

			ID: m.ID,

			OwnedBy:   "openai",
			CreatedAt: time.Now().Unix(),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (s *Server) handleModel(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	model, found := s.Model(id)

	if !found {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	result := openai.Model{
		ID: model.ID,

		Object:    "model",
		CreatedAt: time.Now().Unix(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
