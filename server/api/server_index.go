package api

import (
	"encoding/json"
	"net/http"

	"github.com/adrianliechti/llama/pkg/index"

	"github.com/go-chi/chi/v5"
)

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	i, err := s.Index(chi.URLParam(r, "index"))

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

			Content:  d.Content,
			Metadata: d.Metadata,
		}

		documents = append(documents, document)
	}

	if err := i.Index(r.Context(), documents...); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleIndexQuery(w http.ResponseWriter, r *http.Request) {
	i, err := s.Index(chi.URLParam(r, "index"))

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var query Query

	if err := json.NewDecoder(r.Body).Decode(&query); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if len(query.Text) == 0 && len(query.Embedding) == 0 {
		writeError(w, http.StatusBadRequest, nil)
		return
	}

	if query.Embedding == nil {
		embedding, err := i.Embed(r.Context(), query.Text)

		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		query.Embedding = embedding
	}

	options := &index.QueryOptions{
		Limit:    query.Limit,
		Distance: query.Distance,
	}

	result, err := i.Query(r.Context(), query.Embedding, options)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	results := make([]Result, 0)

	for _, r := range result {
		results = append(results, Result{
			Document: Document{
				ID: r.ID,

				Content:  r.Content,
				Metadata: r.Metadata,
			},

			Distance: r.Distance,
		})
	}

	writeJson(w, results)
}
