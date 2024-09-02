package ollama

import (
	"bytes"
	"encoding/json"
	"fmt"
	"mime"
	"net/http"
	"time"

	"github.com/adrianliechti/llama/pkg/provider"
)

func (h *Handler) handleChat(w http.ResponseWriter, r *http.Request) {
	var req ChatRequest

	if req.Stream == nil {
		val := true
		req.Stream = &val
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	completer, err := h.Completer(req.Model)

	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	messages, err := toMessages(req.Messages)

	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	options := &provider.CompleteOptions{}

	if *req.Stream {
		w.Header().Set("Content-Type", "application/x-ndjson")

		done := make(chan error)
		stream := make(chan provider.Completion)

		go func() {
			options.Stream = stream

			completion, err := completer.Complete(r.Context(), messages, options)

			select {
			case <-stream:
				break
			default:
				if completion != nil {
					stream <- *completion
				}

				close(stream)
			}

			done <- err
		}()

		for completion := range stream {
			timestamp := time.Now().UTC()

			role := ollamaMessageRole(completion.Message.Role)
			content := completion.Message.Content

			if role == "" {
				role = MessageRoleAssistant
			}

			result := ChatResponse{
				Model: req.Model,

				CreatedAt: timestamp,

				Message: Message{
					Role:    role,
					Content: content,
				},

				Done: completion.Reason == provider.CompletionReasonStop,
			}

			var data bytes.Buffer

			enc := json.NewEncoder(&data)
			enc.SetEscapeHTML(false)
			enc.Encode(result)

			fmt.Fprintf(w, "%s\n", data.String())
			w.(http.Flusher).Flush()
		}

		if err := <-done; err != nil {
			writeError(w, http.StatusBadRequest, err)
		}

	} else {
		completion, err := completer.Complete(r.Context(), messages, options)

		if err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}

		timestamp := time.Now().UTC()

		role := ollamaMessageRole(completion.Message.Role)
		content := completion.Message.Content

		if role == "" {
			role = MessageRoleAssistant
		}

		result := ChatResponse{
			Model: req.Model,

			CreatedAt: timestamp,

			Message: Message{
				Role:    role,
				Content: content,
			},

			Done: completion.Reason == provider.CompletionReasonStop,
		}

		writeJson(w, result)
	}
}

func toMessages(s []Message) ([]provider.Message, error) {
	result := make([]provider.Message, 0)

	for _, m := range s {
		files := make([]provider.File, 0)

		for i, data := range m.Images {
			var name string

			if ext, _ := mime.ExtensionsByType(http.DetectContentType(data)); len(ext) > 0 {
				name = fmt.Sprintf("image%03d%s", i+1, ext[0])
			}

			if name == "" {
				return nil, fmt.Errorf("invalid image data")
			}

			files = append(files, provider.File{
				Name:    name,
				Content: bytes.NewReader(data),
			})
		}

		result = append(result, provider.Message{
			Role:    toMessageRole(m.Role),
			Content: m.Content,

			Files: files,
		})
	}

	return result, nil
}

func toMessageRole(r MessageRole) provider.MessageRole {
	switch r {
	case MessageRoleSystem:
		return provider.MessageRoleSystem

	case MessageRoleUser:
		return provider.MessageRoleUser

	case MessageRoleAssistant:
		return provider.MessageRoleAssistant

	default:
		return ""
	}
}

func ollamaMessageRole(r provider.MessageRole) MessageRole {
	switch r {
	case provider.MessageRoleSystem:
		return MessageRoleSystem

	case provider.MessageRoleUser:
		return MessageRoleUser

	case provider.MessageRoleAssistant:
		return MessageRoleAssistant

	default:
		return ""
	}
}
