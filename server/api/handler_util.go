package api

import (
	"io"
	"mime"
	"net/http"
	"strings"

	"github.com/adrianliechti/llama/pkg/extractor"
)

func valueModel(r *http.Request) string {
	if val := r.FormValue("model"); val != "" {
		return val
	}

	return ""
}

func valueURL(r *http.Request) string {
	if val := r.FormValue("url"); val != "" {
		return val
	}

	return ""
}

func valueLanguage(r *http.Request) string {
	if val := r.FormValue("lang"); val != "" {
		return val
	}

	if val := r.FormValue("language"); val != "" {
		return val
	}

	return ""
}

func (h *Handler) readContent(r *http.Request) (string, io.ReadCloser, error) {
	e, err := h.Extractor("")

	if err != nil {
		return "", nil, err
	}

	input := extractor.File{
		URL: r.FormValue("url"),
	}

	if input.URL == "" {
		name, reader, err := h.readFile(r)

		if err != nil {
			return "", nil, err
		}

		input.Name = name
		input.Reader = reader
	}

	document, err := e.Extract(r.Context(), input, nil)

	if err != nil {
		return "", nil, err
	}

	return "file.txt", io.NopCloser(strings.NewReader(document.Content)), nil
}

func (h *Handler) readFile(r *http.Request) (string, io.ReadCloser, error) {
	contentType := r.Header.Get("Content-Type")
	contentDisposition := r.Header.Get("Content-Disposition")

	if strings.Contains(contentType, "multipart/form-data") || strings.Contains(contentType, "application/x-www-form-urlencoded") {
		if file, header, err := r.FormFile("file"); err == nil {
			return header.Filename, file, nil
		}

		if file, header, err := r.FormFile("files"); err == nil {
			return header.Filename, file, nil
		}

		if file, header, err := r.FormFile("input"); err == nil {
			return header.Filename, file, nil
		}
	}

	_, params, _ := mime.ParseMediaType(contentDisposition)

	filename := params["filename*"]
	filename = strings.TrimPrefix(filename, "UTF-8''")
	filename = strings.TrimPrefix(filename, "utf-8''")

	if filename == "" {
		filename = params["filename"]
	}

	return filename, r.Body, nil
}
