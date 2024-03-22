package oai

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/adrianliechti/llama/pkg/provider"
)

func (s *Server) handleAudioSpeech(w http.ResponseWriter, r *http.Request) {
	var req SpeechRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	synthesizer, err := s.Synthesizer(req.Model)

	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	options := &provider.SynthesizeOptions{
		Voice: req.Voice,
	}

	synthesis, err := synthesizer.Synthesize(r.Context(), req.Input, options)

	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	defer synthesis.Content.Close()

	w.Header().Set("Content-Type", "audio/wav")
	io.Copy(w, synthesis.Content)
}
