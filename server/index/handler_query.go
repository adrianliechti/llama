package index

import (
	"encoding/json"
	"net/http"

	"github.com/adrianliechti/llama/pkg/index"
	"github.com/adrianliechti/llama/pkg/to"
)

func (s *Handler) handleQuery(w http.ResponseWriter, r *http.Request) {
	i, err := s.Index(r.PathValue("index"))

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var query Query

	if err := json.NewDecoder(r.Body).Decode(&query); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if len(query.Text) == 0 {
		writeError(w, http.StatusBadRequest, nil)
		return
	}

	options := &index.QueryOptions{
		Limit: query.Limit,
	}

	result, err := i.Query(r.Context(), query.Text, options)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	results := make([]Result, 0)

	for _, r := range result {
		results = append(results, Result{
			Score: to.Ptr(float64(r.Score)),

			Document: Document{
				ID: r.ID,

				Content:  r.Content,
				Metadata: r.Metadata,
			},
		})
	}

	writeJson(w, results)
}
