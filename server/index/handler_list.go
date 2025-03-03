package index

import (
	"net/http"

	"github.com/adrianliechti/wingman/pkg/index"
)

func (s *Handler) handleList(w http.ResponseWriter, r *http.Request) {
	i, err := s.Index(r.PathValue("index"))

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	opts := &index.ListOptions{}

	if val := r.URL.Query().Get("cursor"); val != "" {
		opts.Cursor = val
	}

	page, err := i.List(r.Context(), opts)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	items := make([]Document, 0)

	for _, d := range page.Items {
		items = append(items, Document{
			ID: d.ID,

			Title:   d.Title,
			Source:  d.Source,
			Content: d.Content,

			Metadata: d.Metadata,

			Embedding: d.Embedding,
		})
	}

	result := Page[Document]{
		Items:  items,
		Cursor: page.Cursor,
	}

	writeJson(w, result)
}
