package server

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/sashabaranov/go-openai"
)

func (s *Server) handleEmbeddings(w http.ResponseWriter, r *http.Request) {
	var req openai.EmbeddingRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	model := req.Model.String()
	inputs, err := convertEmbeddingInputs(req)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	result := &openai.EmbeddingResponse{
		Object: "list",

		Model: req.Model,
	}

	for i, input := range inputs {
		data, err := s.provider.Embed(r.Context(), model, input)

		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		result.Data = append(result.Data, openai.Embedding{
			Object: "embedding",

			Index:     i,
			Embedding: data,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func convertEmbeddingInputs(req openai.EmbeddingRequest) ([]string, error) {
	data, _ := json.Marshal(req)

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
