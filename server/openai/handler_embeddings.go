package openai

import (
	"encoding/json"
	"errors"
	"net/http"
)

func (h *Handler) handleEmbeddings(w http.ResponseWriter, r *http.Request) {
	var req EmbeddingsRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	embedder, err := h.Embedder(req.Model)

	if err != nil {
		writeError(w, http.StatusBadRequest, err)
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
		writeError(w, http.StatusBadRequest, errors.New("no input provided"))
		return
	}

	result := &EmbeddingList{
		Object: "list",

		Model: req.Model,
	}

	inputTokens := 0
	outputToken := 0

	for i, input := range inputs {
		embedding, err := embedder.Embed(r.Context(), input)

		if err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}

		result.Data = append(result.Data, Embedding{
			Object: "embedding",

			Index:     i,
			Embedding: embedding.Data,
		})

		if embedding.Usage != nil {
			inputTokens += embedding.Usage.InputTokens
			outputToken += embedding.Usage.OutputTokens
		}
	}

	if inputTokens > 0 || outputToken > 0 {
		result.Usage = &Usage{
			PromptTokens: inputTokens,
			TotalTokens:  inputTokens + outputToken,
		}
	}

	writeJson(w, result)
}
