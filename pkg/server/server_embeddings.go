package server

import (
	"encoding/json"
	"net/http"

	"github.com/sashabaranov/go-openai"
)

func (s *Server) handleEmbeddings(w http.ResponseWriter, r *http.Request) {
	var req openai.EmbeddingRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	result, err := s.provider.Embed(r.Context(), req)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
