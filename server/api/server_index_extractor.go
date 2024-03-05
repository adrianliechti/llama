package api

import (
	"fmt"
	"net/http"

	"github.com/adrianliechti/llama/pkg/extractor"
	"github.com/adrianliechti/llama/pkg/index"

	"github.com/go-chi/chi/v5"
)

func (s *Server) handleIndexWithExtractor(w http.ResponseWriter, r *http.Request) {
	i, err := s.Index(chi.URLParam(r, "index"))

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	e, err := s.Extractor(chi.URLParam(r, "extractor"))

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	file := extractor.File{
		Name:    detectFileName(r),
		Content: r.Body,
	}

	if file.Name == "" {
		http.Error(w, "invalid content type", http.StatusBadRequest)
		return
	}

	data, err := e.Extract(r.Context(), file, nil)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var documents []index.Document

	for i, p := range data.Blocks {
		document := index.Document{
			ID:      fmt.Sprintf("%s#%d", file.Name, i),
			Content: p.Content,

			Metadata: map[string]string{
				"filename": file.Name,
				"filepart": fmt.Sprintf("%d", i),
			},
		}

		documents = append(documents, document)
	}

	if err := i.Index(r.Context(), documents...); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
