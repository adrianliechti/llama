package oai

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

func (s *Server) handleModels(w http.ResponseWriter, r *http.Request) {
	result := &ModelList{
		Object: "list",
	}

	for _, m := range s.Models() {
		result.Models = append(result.Models, Model{
			Object: "model",

			ID:      m.ID,
			Created: time.Now().Unix(),
			OwnedBy: "openai",
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (s *Server) handleModel(w http.ResponseWriter, r *http.Request) {
	model, err := s.Model(chi.URLParam(r, "id"))

	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	result := &Model{
		Object: "model",

		ID:      model.ID,
		Created: time.Now().Unix(),
		OwnedBy: "openai",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
