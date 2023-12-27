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
	messages := []provider.CompletionMessage{}

	if prompt, ok := req.Prompt.(string); ok {
		messages = []provider.CompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		}
	}

	if prompts, ok := req.Prompt.([]string); ok {
		messages = []provider.CompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: strings.Join(prompts, "\n"),
			},
		}
	}

	if req.Stream {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		done := make(chan error)
		stream := make(chan provider.Completion)

		go func() {
			done <- s.provider.CompleteStream(r.Context(), model, messages, stream)
		}()

		for {
			select {
			case result := <-stream:
				status := ""

				if result.Result == provider.MessageResultStop {
					status = "stop"
				}

				completion := openai.CompletionResponse{
					ID: id,

					Object:  "text_completion",
					Created: time.Now().Unix(),

					Model: model,

					Choices: []openai.CompletionChoice{
						{
							Text: result.Message.Content,

							FinishReason: status,
						},
					},
				}

				data, _ := json.Marshal(completion)

				fmt.Fprintf(w, "data: %s\n\n", string(data))
				w.(http.Flusher).Flush()

			case err := <-done:
				time.Sleep(1 * time.Second)

				fmt.Fprintf(w, "data: [DONE]\n\n")
				w.(http.Flusher).Flush()

				if err != nil {
					slog.Error("error in completion", "error", err)
				}

				return

			case <-r.Context().Done():
				return
			}
		}
	} else {
		result, err := s.provider.Complete(r.Context(), model, messages)

		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		status := ""

		if result.Result == provider.MessageResultStop {
			status = "stop"
		}

		completion := openai.CompletionResponse{
			ID: id,

			Object:  "text_completion",
			Created: time.Now().Unix(),

			Model: model,

			Choices: []openai.CompletionChoice{
				{
					Text: result.Message.Content,

					FinishReason: status,
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(completion)
	}
}
