package server

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/adrianliechti/llama/pkg/provider"

	"github.com/google/uuid"
	"github.com/sashabaranov/go-openai"
)

func (s *Server) handleCompletions(w http.ResponseWriter, r *http.Request) {
	var req openai.CompletionRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	id := uuid.New().String()

	model := req.Model
	messages := []provider.Message{}

	if prompt, ok := req.Prompt.(string); ok {
		messages = []provider.Message{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		}
	}

	if prompts, ok := req.Prompt.([]string); ok {
		messages = []provider.Message{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: strings.Join(prompts, "\n"),
			},
		}
	}

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
				completion := openai.CompletionResponse{
					ID: id,

					Object:  "text_completion",
					Created: time.Now().Unix(),

					Model: model,

					Choices: []openai.CompletionChoice{
						{
							Text: message.Content,
						},
					},
				}

				data, _ := json.Marshal(completion)

				fmt.Fprintf(w, "data: %s\n\n", string(data))
				w.(http.Flusher).Flush()

			case err := <-done:
				if err != nil {
					slog.Error("error in completion", "error", err)
				}

				completion := openai.CompletionResponse{
					ID: id,

					Object:  "text_completion",
					Created: time.Now().Unix(),

					Model: model,

					Choices: []openai.CompletionChoice{
						{
							Text:         "",
							FinishReason: string(openai.FinishReasonStop),
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

		completion := openai.CompletionResponse{
			ID: id,

			Object:  "text_completion",
			Created: time.Now().Unix(),

			Model: model,

			Choices: []openai.CompletionChoice{
				{
					Text: message.Content,
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(completion)
	}
}
