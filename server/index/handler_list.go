package index

import (
	"net/http"
)

func (s *Handler) handleList(w http.ResponseWriter, r *http.Request) {
	i, err := s.Index(r.PathValue("index"))

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	result, err := i.List(r.Context(), nil)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	results := make([]Document, 0)

	for _, r := range result {
		results = append(results, Document{
			ID: r.ID,

			Title:  r.Title,
			Source: r.Source,
			//Content: r.Content,

			Metadata: r.Metadata,

			//Embedding: r.Embedding,
		})
	}

	writeJson(w, results)
}
