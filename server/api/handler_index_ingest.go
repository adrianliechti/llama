package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/adrianliechti/llama/pkg/extractor"
	"github.com/adrianliechti/llama/pkg/index"
)

func (s *Handler) handleIngest(w http.ResponseWriter, r *http.Request) {
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

func (s *Handler) handleIngestWithExtractor(w http.ResponseWriter, r *http.Request) {
	i, err := s.Index(r.PathValue("index"))

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	e, err := s.Extractor(r.PathValue("extractor"))

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
			ID: fmt.Sprintf("%s#%d", file.Name, i),

			Title:    file.Name,
			Location: fmt.Sprintf("file.Name#%d", i),

			Content:  p.Content,
			Metadata: map[string]string{},
		}

		documents = append(documents, document)
	}

	if err := i.Index(r.Context(), documents...); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
