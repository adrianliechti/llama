package oai

import (
	"encoding/json"
	"net/http"
)

func (s *Server) handleEmbeddings(w http.ResponseWriter, r *http.Request) {
	var req EmbeddingsRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	embedder, err := s.Embedder(req.Model)

	if err != nil {
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

	result := &EmbeddingList{
		Object: "list",

		Model: req.Model,
	}

	for i, input := range inputs {
		data, err := embedder.Embed(r.Context(), input)

		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		result.Data = append(result.Data, Embedding{
			Object: "embedding",

			Index:     i,
			Embedding: data,
		})
	}

	writeJson(w, result)
}
