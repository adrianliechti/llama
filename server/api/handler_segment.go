package api

import (
	"net/http"

	"github.com/adrianliechti/llama/pkg/segmenter"
)

func (h *Handler) handleSegment(w http.ResponseWriter, r *http.Request) {
	model := valueModel(r)

	p, err := h.Segmenter(model)

	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	name, reader, err := h.readContent(r)

	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	input := segmenter.File{
		Name:   name,
		Reader: reader,
	}

	options := &segmenter.SegmentOptions{}

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
