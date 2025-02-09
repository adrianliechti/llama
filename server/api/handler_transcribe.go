package api

import (
	"net/http"

	"github.com/adrianliechti/llama/pkg/provider"
)

func (h *Handler) handleTranscribe(w http.ResponseWriter, r *http.Request) {
	model := r.FormValue("model")
	language := r.FormValue("language")

	file, header, err := r.FormFile("file")

	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	p, err := h.Transcriber(model)

	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	input := provider.File{
		Name: header.Filename,

		Content:     file,
		ContentType: header.Header.Get("Content-Type"),
	}

	options := &provider.TranscribeOptions{
		Language: language,
	}

	transcription, err := p.Transcribe(r.Context(), input, options)

	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	result := Document{
		Content: transcription.Content,
	}

	writeJson(w, result)
}
