package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (s *Server) handleIndexList(w http.ResponseWriter, r *http.Request) {
	i, err := s.Index(chi.URLParam(r, "index"))

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

			Content:  r.Content,
			Metadata: r.Metadata,
		})
	}

	writeJson(w, results)
}
