package index

import (
	"encoding/json"
	"net/http"

	"github.com/adrianliechti/wingman/pkg/index"
)

func (s *Handler) handleIndex(w http.ResponseWriter, r *http.Request) {
	i, err := s.Index(r.PathValue("index"))

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var request []Document

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var documents []index.Document

	for _, d := range request {
		document := index.Document{
			ID: d.ID,

			Title:   d.Title,
			Source:  d.Source,
			Content: d.Content,

			Metadata: d.Metadata,

			Embedding: d.Embedding,
		}

		documents = append(documents, document)
	}

	if err := i.Index(r.Context(), documents...); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
