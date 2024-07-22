package api

import (
	"net/http"

	"github.com/adrianliechti/llama/pkg/extractor"
)

func (h *Handler) handleExtract(w http.ResponseWriter, r *http.Request) {
	e, err := h.Extractor(r.PathValue("extractor"))

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	file := extractor.File{
		Name:    detectFileName(r),
		Content: r.Body,
	}

	if file.Name == "" {
		http.Error(w, "invalid content type", http.StatusBadRequest)
		return
	}

	data, err := e.Extract(r.Context(), file, nil)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var result []Document

	for _, b := range data.Blocks {
		metadata := map[string]string{}

		if data.Name != "" {
			metadata["filename"] = data.Name
		}

		document := Document{
			ID:       b.ID,
			Metadata: metadata,

			Content: b.Content,
		}

		result = append(result, document)
	}

	writeJson(w, result)
}
