package unstructured

import (
	"net/http"

	"github.com/adrianliechti/llama/pkg/partitioner"
)

func (h *Handler) handlePartition(w http.ResponseWriter, r *http.Request) {
	p, err := h.Partitioner("default")

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

	input := partitioner.File{
		Content: file,
		Name:    header.Filename,
	}

	if input.Name == "" {
		http.Error(w, "invalid content type", http.StatusBadRequest)
		return
	}

	partitions, err := p.Partition(r.Context(), input, nil)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var result []Partition

	for _, p := range partitions {
		partition := Partition{
			ID: p.ID,

			Text: p.Content,
		}

		result = append(result, partition)
	}

	writeJson(w, result)
}
