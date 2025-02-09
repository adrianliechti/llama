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

	text := req.Text

	if text == "" {
		text = req.Content
	}

	input := segmenter.File{
		Name:   "file.txt",
		Reader: strings.NewReader(req.Text),
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
