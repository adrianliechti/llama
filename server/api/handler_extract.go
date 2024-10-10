package api

import (
	"errors"
	"io"
	"net/http"

	"github.com/adrianliechti/llama/pkg/extractor"
)

func (h *Handler) handleExtract(w http.ResponseWriter, r *http.Request) {
	model := r.FormValue("model")
	format := r.FormValue("format")

	input := extractor.File{
		URL: r.FormValue("url"),
	}

	if file, header, err := r.FormFile("file"); err == nil {
		input.Name = header.Filename
		input.Content = file
	}

	if input.Content == nil && input.URL == "" {
		writeError(w, http.StatusBadRequest, errors.New("invalid input"))
		return
	}

	p, err := h.Extractor(model)

	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	document, err := p.Extract(r.Context(), input, nil)

	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	_ = format
	//w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Content-Type", "application/octet-stream")

	io.WriteString(w, document.Content)
}
