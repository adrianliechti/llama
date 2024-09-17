package jina

import (
	"encoding/json"
	"net/http"

	"github.com/adrianliechti/llama/pkg/text"
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

	splitter := text.NewSplitter()
	splitter.ChunkSize = req.MaxChunkLength

	chuks := splitter.Split(req.Content)

	result := SegmentResponse{
		Chunks: chuks,
	}

	writeJson(w, result)
}
