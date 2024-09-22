package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/adrianliechti/llama/pkg/provider"
	"github.com/adrianliechti/llama/pkg/segmenter"
	"github.com/adrianliechti/llama/pkg/text"
	"github.com/adrianliechti/llama/pkg/to"
)

func (h *Handler) handleSummarize(w http.ResponseWriter, r *http.Request) {
	var req SummarizeRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	c, err := h.Completer(req.Model)

	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	s, err := h.Segmenter("")

	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	text := text.Normalize(req.Content)

	input := segmenter.File{
		Name:    "input.txt",
		Content: strings.NewReader(text),
	}

	segments, err := s.Segment(r.Context(), input, &segmenter.SegmentOptions{
		SegmentLength: to.Ptr(16000),
	})

	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	var parts []string

	for _, segment := range segments {
		completion, err := c.Complete(r.Context(), []provider.Message{
			{
				Role:    provider.MessageRoleUser,
				Content: "Write a concise summary of the following: \n" + segment.Content,
			},
		}, nil)

		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}

		parts = append(parts, completion.Message.Content)
	}

	completion, err := c.Complete(r.Context(), []provider.Message{
		{
			Role:    provider.MessageRoleUser,
			Content: "Distill the following parts into a consolidated summary: \n" + strings.Join(parts, "\n\n"),
		},
	}, nil)

	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	result := Document{
		Content: completion.Message.Content,
	}

	for _, p := range parts {
		segment := Segment{
			Text: p,
		}

		result.Segements = append(result.Segements, segment)
	}

	writeJson(w, result)
}
