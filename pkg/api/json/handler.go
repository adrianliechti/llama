package json

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/adrianliechti/llama/pkg/api"
	"github.com/adrianliechti/llama/pkg/provider"
)

var _ api.Provider = (*Handler)(nil)

type Handler struct {
	input  *api.Schema
	output *api.Schema

	completer provider.Completer
}

func New(options ...Option) (*Handler, error) {
	h := &Handler{}

	for _, option := range options {
		option(h)
	}

	return h, nil
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var body map[string]any

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	messages := []provider.Message{}

	var system strings.Builder

	if h.input != nil {
		system.WriteString("## Input (`" + h.input.Name + "`):\n")
		system.WriteString(h.input.Description)
		system.WriteString("\n\n")

		schema, _ := json.MarshalIndent(h.input.Schema, "", "  ")

		system.WriteString("Input Schema:")
		system.WriteString("\n```json\n")
		system.WriteString(string(schema))
		system.WriteString("\n```\n\n")
	}

	if h.output != nil {
		system.WriteString("## Output (`" + h.output.Name + "`):\n")
		system.WriteString(h.output.Description)
		system.WriteString("\n\n")

		schema, _ := json.MarshalIndent(h.output.Schema, "", "  ")

		system.WriteString("Output Schema:")
		system.WriteString("\n```json\n")
		system.WriteString(string(schema))
		system.WriteString("\n```\n\n")
	}

	messages = append(messages, provider.Message{
		Role:    provider.MessageRoleSystem,
		Content: system.String(),
	})

	println(system.String())

	input, _ := json.MarshalIndent(body, "", "  ")

	messages = append(messages, provider.Message{
		Role:    provider.MessageRoleUser,
		Content: string(input),
	})

	options := &provider.CompleteOptions{
		Format: provider.CompletionFormatJSON,
		Schema: h.output,
	}

	completion, err := h.completer.Complete(r.Context(), messages, options)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := completion.Message.Content

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(data))
}
