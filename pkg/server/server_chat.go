package server

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/adrianliechti/llama/pkg/provider"
	"github.com/sashabaranov/go-openai"

	"github.com/google/uuid"
)

func (s *Server) handleChatCompletions(w http.ResponseWriter, r *http.Request) {
	var req openai.ChatCompletionRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	id := uuid.New().String()

	model := req.Model
	messages := convertCompletionMessages(req.Messages)

	options := &provider.CompleteOptions{}

	if req.Stream {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		done := make(chan error)
		stream := make(chan provider.Message)

		go func() {
			done <- s.provider.CompleteStream(r.Context(), model, messages, stream, options)
		}()

		for {
			select {
			case message := <-stream:
				completion := openai.ChatCompletionStreamResponse{
					ID: id,

					Object:  "chat.completion.chunk",
					Created: time.Now().Unix(),

					Model: model,

					Choices: []openai.ChatCompletionStreamChoice{
						{
							Delta: openai.ChatCompletionStreamChoiceDelta{
								Role:    toMessageRole(message.Role),
								Content: message.Content,
							},
						},
					},
				}

				data, _ := json.Marshal(completion)

				fmt.Fprintf(w, "data: %s\n\n", string(data))
				w.(http.Flusher).Flush()

			case err := <-done:
				if err != nil {
					slog.Error("error in chat completion", "error", err)
				}

				completion := openai.ChatCompletionStreamResponse{
					ID: id,

					Object:  "chat.completion.chunk",
					Created: time.Now().Unix(),

					Model: model,

					Choices: []openai.ChatCompletionStreamChoice{
						{
							Delta: openai.ChatCompletionStreamChoiceDelta{
								Role:    toMessageRole(provider.MessageRoleAssistant),
								Content: "",
							},

							FinishReason: openai.FinishReasonStop,
						},
					},
				}

				data, _ := json.Marshal(completion)

				fmt.Fprintf(w, "data: %s\n\n", string(data))
				w.(http.Flusher).Flush()

				fmt.Fprintf(w, "data: [DONE]\n\n")
				w.(http.Flusher).Flush()

				return

			case <-r.Context().Done():
				return
			}
		}
	} else {
		message, err := s.provider.Complete(r.Context(), model, messages, options)

		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		completion := openai.ChatCompletionResponse{
			ID: id,

			Object:  "chat.completion",
			Created: time.Now().Unix(),

			Model: model,

			Choices: []openai.ChatCompletionChoice{
				{
					Message: openai.ChatCompletionMessage{
						Role:    toMessageRole(message.Role),
						Content: message.Content,
					},
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(completion)
	}
}

func convertCompletionMessages(s []openai.ChatCompletionMessage) []provider.Message {
	result := make([]provider.Message, 0)

	for _, m := range s {
		result = append(result, convertCompletionMessage(m))
	}

	return result
}

func convertCompletionMessage(m openai.ChatCompletionMessage) provider.Message {
	return provider.Message{
		Role:    convertMessageRole(m.Role),
		Content: m.Content,
	}
}

func convertMessageRole(r string) provider.MessageRole {
	switch r {
	case openai.ChatMessageRoleSystem:
		return provider.MessageRoleSystem

	case openai.ChatMessageRoleUser:
		return provider.MessageRoleUser

	case openai.ChatMessageRoleAssistant:
		return provider.MessageRoleAssistant

	// case openai.ChatMessageRoleFunction:
	// 	return provider.MessageRoleFunction

	// case openai.ChatMessageRoleTool:
	// 	return provider.MessageRoleTool

	default:
		return ""
	}
}

func toMessageRole(r provider.MessageRole) string {
	switch r {
	case provider.MessageRoleSystem:
		return openai.ChatMessageRoleSystem

	case provider.MessageRoleUser:
		return openai.ChatMessageRoleUser

	case provider.MessageRoleAssistant:
		return openai.ChatMessageRoleAssistant

	default:
		return ""
	}
}
