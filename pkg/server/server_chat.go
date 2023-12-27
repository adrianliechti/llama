package server

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/adrianliechti/llama/pkg/provider"

	"github.com/adrianliechti/llama/pkg/server/models"
	"github.com/adrianliechti/llama/pkg/server/models/convert"
	"github.com/adrianliechti/llama/pkg/server/models/to"

	"github.com/google/uuid"
)

func (s *Server) handleChatCompletions(w http.ResponseWriter, r *http.Request) {
	var req models.ChatCompletionRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	id := uuid.New().String()

	model := req.Model
	messages := to.CompletionMessages(req.Messages)

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
				completion := models.ChatCompletion{
					ID: id,

					Object:  "chat.completion.chunk",
					Created: time.Now().Unix(),

					Model: model,

					Choices: []models.ChatCompletionChoice{
						{
							Delta: models.ChatCompletionMessage{
								Role:    convert.MessageRole(result.Message.Role),
								Content: result.Message.Content,
							},

							FinishReason: convert.MessageResult(result.Result),
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
					slog.Error("error in chat completion", "error", err)
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

		completion := models.ChatCompletion{
			ID: id,

			Object:  "chat.completion",
			Created: time.Now().Unix(),

			Model: model,

			Choices: []models.ChatCompletionChoice{
				{
					Message: models.ChatCompletionMessage{
						Role:    convert.MessageRole(result.Message.Role),
						Content: result.Message.Content,
					},

					FinishReason: convert.MessageResult(result.Result),
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(completion)
	}
}
