package api

import (
	"io"
	"net/http"

	"github.com/adrianliechti/llama/pkg/translator"
)

func (h *Handler) handleTranslate(w http.ResponseWriter, r *http.Request) {
	model := valueModel(r)
	language := valueLanguage(r)

	p, err := h.Translator(model)

	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	_, reader, err := h.readContent(r)

	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	defer reader.Close()

	data, err := io.ReadAll(reader)

	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	options := &translator.TranslateOptions{
		Language: language,
	}

	translation, err := p.Translate(r.Context(), string(data), options)

	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	w.Header().Set("Content-Type", "text/plain")

	w.WriteHeader(http.StatusOK)
	io.WriteString(w, translation.Content)
}
