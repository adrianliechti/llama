package oai

import (
	"net/http"

	"github.com/adrianliechti/llama/pkg/provider"
)

func (s *Server) handleAudioTranscriptions(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	model := r.FormValue("model")

	transcriber, err := s.Transcriber(model)

	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	prompt := r.FormValue("prompt")
	language := r.FormValue("language")

	_ = prompt
	_ = language

	file, header, err := r.FormFile("file")

	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	options := &provider.TranscribeOptions{
		Name: header.Filename,
	}

	transcription, err := transcriber.Transcribe(r.Context(), file, options)

	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	result := Transcription{
		Text: transcription.Content,
	}

	writeJson(w, result)
}
