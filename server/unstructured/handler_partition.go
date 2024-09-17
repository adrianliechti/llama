package unstructured

import (
	"fmt"
	"net/http"

	"github.com/adrianliechti/llama/pkg/extractor"
	"github.com/adrianliechti/llama/pkg/text"
)

func (h *Handler) handlePartition(w http.ResponseWriter, r *http.Request) {
	e, err := h.Extractor("default")

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
		splitter := text.NewSplitter()
		chunks := splitter.Split(document.Content)

		result = []Partition{}

		for i, chunk := range chunks {
			partition := Partition{
				ID:   fmt.Sprintf("%s#%d", input.Name, i),
				Text: chunk,
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
