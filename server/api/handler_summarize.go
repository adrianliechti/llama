package api

import (
	"encoding/json"
	"net/http"
)

func (h *Handler) handleSummarize(w http.ResponseWriter, r *http.Request) {
	var req SummarizeRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	p, err := h.Summarizer(req.Model)

	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	summary, err := p.Summarize(r.Context(), req.Content, nil)

	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	result := Document{
		Content: summary.Text,
	}

	for _, s := range summary.Segments {
		segment := Segment{
			Text: s,
		}

		result.Segements = append(result.Segements, segment)
	}

	writeJson(w, result)
}
