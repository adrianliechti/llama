package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/adrianliechti/llama/pkg/segmenter"
)

func (h *Handler) handleSegment(w http.ResponseWriter, r *http.Request) {
	var req SegmentRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	p, err := h.Segmenter("")

	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	input := segmenter.File{
		Name:    "file.txt",
		Content: strings.NewReader(req.Content),
	}

	options := &segmenter.SegmentOptions{
		SegmentLength:  req.SegmentLength,
		SegmentOverlap: req.SegmentOverlap,
	}

	segments, err := p.Segment(r.Context(), input, options)

	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	result := Document{}

	for _, s := range segments {
		segment := Segment{
			Text: s.Content,
		}

		result.Segments = append(result.Segments, segment)
	}

	writeJson(w, result)
}
