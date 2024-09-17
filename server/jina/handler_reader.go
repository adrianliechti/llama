package jina

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/adrianliechti/llama/pkg/extractor"
)

func (h *Handler) handleRead(w http.ResponseWriter, r *http.Request) {
	var req ReadRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	format := r.Header.Get("X-Return-Format")
	_ = format

	e, err := h.Extractor("")

	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	input, err := convertInput(req)

	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	options, err := convertOptions(req)

	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	document, err := e.Extract(r.Context(), *input, options)

	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	w.Header().Set("Content-Type", "text/plain")

	io.WriteString(w, document.Content)
}

func convertInput(req ReadRequest) (*extractor.File, error) {
	f := &extractor.File{
		URL: req.URL,
	}

	if req.PDF != "" {
		data, err := base64.StdEncoding.DecodeString(req.PDF)

		if err != nil {
			return nil, err
		}

		f.Name = "file.pdf"
		f.Content = bytes.NewReader(data)
	}

	if req.HTML != "" {
		data, err := base64.StdEncoding.DecodeString(req.HTML)

		if err != nil {
			return nil, err
		}

		f.Name = "file.html"
		f.Content = bytes.NewReader(data)
	}

	if f.URL == "" && f.Content == nil {
		return nil, errors.New("invalid input")
	}

	return f, nil
}

func convertOptions(req ReadRequest) (*extractor.ExtractOptions, error) {
	return &extractor.ExtractOptions{}, nil
}
