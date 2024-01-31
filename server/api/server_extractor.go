package api

import (
	"net/http"

	"github.com/adrianliechti/llama/pkg/extracter"

	"github.com/go-chi/chi/v5"
)

func (s *Server) handleExtract(w http.ResponseWriter, r *http.Request) {
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

	var result Document

	for _, p := range data.Pages {
		page := Page{}

		for _, b := range p.Blocks {
			block := Block{
				Content: b.Text,
			}

			page.Blocks = append(page.Blocks, block)
		}

		result.Pages = append(result.Pages, page)
	}

	writeJson(w, result)
}
