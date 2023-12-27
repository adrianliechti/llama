package server

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/adrianliechti/llama/pkg/server/models"
)

func (s *Server) handleEmbeddings(w http.ResponseWriter, r *http.Request) {
	var req models.EmbeddingRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	inputs, err := convertEmbeddingInputs(req)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	result := &models.Embeddings{
		Object: "list",

		Model: req.Model,
	}

	for i, input := range inputs {
		embedding, err := s.provider.Embed(r.Context(), req.Model, input)

		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		result.Data = append(result.Data, models.Embedding{
			Object: "embedding",

			Index:     i,
			Embedding: embedding.Embeddings,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func convertEmbeddingInputs(request models.EmbeddingRequest) ([]string, error) {
	data, _ := json.Marshal(request)

	type stringType struct {
		Input string `json:"input"`
	}

	var stringVal stringType

	if json.Unmarshal(data, &stringVal) == nil {
		if stringVal.Input != "" {
			return []string{stringVal.Input}, nil
		}
	}

	type sliceType struct {
		Input []string `json:"input"`
	}

	var sliceVal sliceType

	if json.Unmarshal(data, &sliceVal) == nil {
		if len(sliceVal.Input) > 0 {
			return sliceVal.Input, nil
		}
	}

	return nil, errors.New("invalid input format")
}
