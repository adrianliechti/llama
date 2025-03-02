package api

import (
	"net/http"

	"github.com/adrianliechti/wingman/pkg/segmenter"
)

func (h *Handler) handleSegment(w http.ResponseWriter, r *http.Request) {
	model := valueModel(r)

	p, err := h.Segmenter(model)

	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	text, err := h.readText(r)

	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	options := &segmenter.SegmentOptions{}

	segments, err := p.Segment(r.Context(), text, options)

	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	result := make([]Segment, 0)

	for _, s := range segments {
		segment := Segment{
			Text: s.Text,
		}

		result = append(result, segment)
	}

	writeJson(w, result)
}
