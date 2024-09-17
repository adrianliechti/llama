package jina

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

	if req.MaxChunkLength < 500 {
		req.MaxChunkLength = 500
	}

	if req.MaxChunkLength > 2000 {
		req.MaxChunkLength = 2000
	}

	s, err := h.Segmenter("")

	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	input := segmenter.File{
		Name:    "input.txt",
		Content: strings.NewReader(req.Content),
	}

	segments, err := s.Segment(r.Context(), input, &segmenter.SegmentOptions{
		SegmentLength: &req.MaxChunkLength,
	})

	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	var chuks []string

	for _, s := range segments {
		chuks = append(chuks, s.Content)
	}

	result := SegmentResponse{
		Chunks: chuks,
	}

	writeJson(w, result)
}
