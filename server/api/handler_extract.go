package api

import (
	"errors"
	"io"
	"net/http"

	"github.com/adrianliechti/llama/pkg/extractor"
)

func (h *Handler) handleExtract(w http.ResponseWriter, r *http.Request) {
	model := valueModel(r)

	p, err := h.Extractor(model)

	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	input := extractor.File{
		URL: valueURL(r),
	}

	if file, header, err := r.FormFile("file"); err == nil {
		input.Name = header.Filename
		input.Reader = file
	}

	if input.URL == "" && input.Reader == nil {
		writeError(w, http.StatusBadRequest, errors.New("invalid input"))
		return
	}

	options := &extractor.ExtractOptions{}

	document, err := p.Extract(r.Context(), input, options)

	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	contentType := document.ContentType

	if contentType != "" {
		contentType = "application/octet-stream"
	}

	w.Header().Set("Content-Type", contentType)

	w.WriteHeader(http.StatusOK)
	io.WriteString(w, document.Content)
}
