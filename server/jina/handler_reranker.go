package jina

import (
	"cmp"
	"encoding/json"
	"net/http"
	"slices"

	"github.com/adrianliechti/llama/pkg/provider"
)

func (h *Handler) handleRerank(w http.ResponseWriter, r *http.Request) {
	var req RerankRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	reranker, err := h.Reranker(req.Model)

	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	rankings, err := reranker.Rerank(r.Context(), req.Query, req.Documents, &provider.RerankOptions{
		Limit: req.TopN,
	})

	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	result := RerankResponse{
		Model: req.Model,
	}

	for _, r := range rankings {
		index := slices.Index(req.Documents, r.Content)

		if index < 0 {
			continue
		}

		result.Results = append(result.Results, RerankResult{
			Index: index,

			Document: Document{
				Text: r.Content,
			},

			RelevanceScore: r.Score,
		})
	}

	slices.SortFunc(result.Results, func(i, j RerankResult) int {
		return cmp.Compare(j.RelevanceScore, i.RelevanceScore)
	})

	writeJson(w, result)
}
