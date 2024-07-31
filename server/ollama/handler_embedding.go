package ollama

import (
	"encoding/json"
	"net/http"
)

func (h *Handler) handleEmbeddings(w http.ResponseWriter, r *http.Request) {
	var req EmbeddingRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	embedder, err := h.Embedder(req.Model)

	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	embedding, err := embedder.Embed(r.Context(), req.Prompt)

	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	result := EmbeddingResponse{
		Embedding: toFloat64s(embedding.Data),
	}

	writeJson(w, result)
}

func toFloat64s(v []float32) []float64 {
	result := make([]float64, len(v))

	for i, x := range v {
		result[i] = float64(x)
	}

	return result
}
