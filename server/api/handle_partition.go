package api

import (
	"net/http"

	"github.com/adrianliechti/llama/pkg/extractor"
)

func (h *Handler) handlePartition(w http.ResponseWriter, r *http.Request) {
	e, err := h.Extractor("default")

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("files")

	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	defer file.Close()

	input := extractor.File{
		Content: file,
		Name:    header.Filename,
	}

	if input.Name == "" {
		http.Error(w, "invalid content type", http.StatusBadRequest)
		return
	}

	data, err := e.Extract(r.Context(), input, nil)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var result []Partition

	for _, b := range data.Blocks {
		partition := Partition{
			ID: b.ID,

			Text: b.Content,

			Metadata: PartitionMetadata{
				FileName: data.Name,
			},
		}

		result = append(result, partition)
	}

	writeJson(w, result)
}
