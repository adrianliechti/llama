package api

import (
	"io"
	"net/http"

	"github.com/adrianliechti/llama/pkg/summarizer"
)

func (h *Handler) handleSummarize(w http.ResponseWriter, r *http.Request) {
	model := valueModel(r)

	p, err := h.Summarizer(model)

	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	text, err := h.readText(r)

	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	options := &summarizer.SummarizerOptions{}

	summary, err := p.Summarize(r.Context(), text, options)

	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	w.Header().Set("Content-Type", "text/plain")

	w.WriteHeader(http.StatusOK)
	io.WriteString(w, summary.Text)
}
