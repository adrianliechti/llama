package oai

import (
	"encoding/json"
	"net/http"

	"github.com/sashabaranov/go-openai"
)

func (s *Server) handleEmbeddings(w http.ResponseWriter, r *http.Request) {
	type embeddingRequest struct {
		Input any    `json:"input"`
		Model string `json:"model"`
	}

	var req embeddingRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	p, found := s.Provider(req.Model)

	if !found {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var inputs []string

	switch v := req.Input.(type) {
	case string:
		inputs = []string{v}
	case []string:
		inputs = v
	}

	if len(inputs) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	result := &openai.EmbeddingResponse{
		Object: "list",

		Model: openai.AdaEmbeddingV2,
	}

	for i, input := range inputs {
		data, err := p.Embed(r.Context(), req.Model, input)

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
