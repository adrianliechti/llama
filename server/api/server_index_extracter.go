package api

import (
	"fmt"
	"net/http"
	"strconv"
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

	var documents []index.Document

	for i, p := range data.Pages {
		var content strings.Builder

		for _, b := range p.Blocks {
			content.WriteString(b.Text)
			content.WriteString("\n")
		}

		page := i + 1

		document := index.Document{
			ID:      fmt.Sprintf("%s#%d", file.Name, page),
			Content: content.String(),

			Metadata: map[string]string{
				"filename": file.Name,
				"page":     strconv.Itoa(page),
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
