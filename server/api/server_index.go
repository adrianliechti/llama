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

func (s *Server) handleIndexSearch(w http.ResponseWriter, r *http.Request) {
	i, err := s.Index(chi.URLParam(r, "index"))

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var request SearchRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if len(request.Embedding) == 0 && len(request.Content) == 0 {
		writeError(w, http.StatusBadRequest, nil)
		return
	}

	if request.Embedding == nil {
		embedding, err := i.Embed(r.Context(), request.Content)

		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		request.Embedding = embedding
	}

	options := &index.SearchOptions{
		TopK: request.TopK,
		TopP: request.TopP,
	}

	result, err := i.Search(r.Context(), request.Embedding, options)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	_ = result

	w.WriteHeader(http.StatusNoContent)
}
