package api

import (
	"net/http"
	"strings"

	"github.com/adrianliechti/llama/pkg/extracter"
	"github.com/adrianliechti/llama/pkg/index"

	"github.com/go-chi/chi/v5"
)

func (s *Server) handleIndexWithExtracter(w http.ResponseWriter, r *http.Request) {
	i, err := s.Index(chi.URLParam(r, "index"))

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	e, err := s.Extracter(chi.URLParam(r, "extracter"))

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	file := extracter.File{
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

	var content strings.Builder

	for _, b := range data.Blocks {
		content.WriteString(b.Text)
		content.WriteString("\n")
	}

	document := index.Document{
		Content: content.String(),

		Metadata: map[string]string{
			"filename": file.Name,
		},
	}

	if err := i.Index(r.Context(), document); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
