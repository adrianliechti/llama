package unstructured

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/adrianliechti/llama/pkg/extractor"
	"github.com/adrianliechti/llama/pkg/segmenter"
)

func (h *Handler) handlePartition(w http.ResponseWriter, r *http.Request) {
	e, err := h.Extractor("")

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	input := extractor.File{
		URL: r.FormValue("url"),
	}

	if input.URL == "" {
		file, header, err := r.FormFile("files")

		if err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}

		if header.Filename == "" {
			http.Error(w, "invalid content type", http.StatusBadRequest)
			return
		}

		defer file.Close()

		input.Name = header.Filename
		input.Content = file
	}

	chunkStrategy := parseChunkingStrategy(r.FormValue("chunking_strategy"))
	chunkLength := 500
	chunkOverlap := 0

	// if chunkStrategy == ChunkingStrategyUnknown {
	// 	http.Error(w, "invalid chunking strategy", http.StatusBadRequest)
	// 	return
	// }

	document, err := e.Extract(r.Context(), input, nil)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	result := []Partition{
		{
			ID:   input.Name,
			Text: document.Content,
		},
	}

	if chunkStrategy != ChunkingStrategyNone {
		s, err := h.Segmenter("")

		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		input := segmenter.File{
			Name:    input.Name,
			Content: strings.NewReader(document.Content),
		}

		segments, err := s.Segment(r.Context(), input, &segmenter.SegmentOptions{
			SegmentLength:  &chunkLength,
			SegmentOverlap: &chunkOverlap,
		})

		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		result = []Partition{}

		for i, s := range segments {
			partition := Partition{
				ID:   fmt.Sprintf("%s#%d", input.Name, i),
				Text: s.Content,
			}

			result = append(result, partition)
		}
	}

	writeJson(w, result)
}

func parseChunkingStrategy(value string) ChunkingStrategy {
	switch value {
	case "none", "":
		return ChunkingStrategyNone
	}

	return ChunkingStrategyUnknown
}
