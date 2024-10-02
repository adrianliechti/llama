package api

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

	p, err := h.Reranker(req.Model)

	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	rankings, err := p.Rerank(r.Context(), req.Query, req.Documents, &provider.RerankOptions{
		Limit: req.Limit,
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

		ranking := Result{
			Index: index,
			Score: r.Score,

			Document: Document{
				Content: r.Content,
			},
		}

		result.Results = append(result.Results, ranking)
	}

	slices.SortFunc(result.Results, func(i, j Result) int {
		return cmp.Compare(j.Score, i.Score)
	})

	writeJson(w, result)
}
