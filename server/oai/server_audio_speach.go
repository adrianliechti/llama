package oai

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/adrianliechti/llama/pkg/provider"
)

func (s *Server) handleAudioSpeech(w http.ResponseWriter, r *http.Request) {
	var req SpeechRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	// HACK: Allow Custom Voices
	if strings.HasPrefix(req.Input, "#") {
		parts := strings.SplitN(req.Input, " ", 2)

		if len(parts) == 2 {
			req.Input = parts[1]
			req.Voice = strings.TrimLeft(parts[0], "#")
		}
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
