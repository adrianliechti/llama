package server

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/sashabaranov/go-openai"
)

func (s *Server) handleCompletions(w http.ResponseWriter, r *http.Request) {
	var req openai.CompletionRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if req.Stream {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		done := make(chan error)
		stream := make(chan openai.ChatCompletionStreamResponse)

		go func() {
			done <- s.provider.CompleteStream(r.Context(), convertCompletionRequest(req), stream)
		}()

		for {
			select {
			case err := <-done:
				fmt.Fprintf(w, "data: [DONE]\n\n")
				w.(http.Flusher).Flush()

				if err != nil {
					slog.Error("error in chat completion", "error", err)
				}

				return

			case resp := <-stream:
				data, _ := json.Marshal(convertChatCompletionStreamResponse(resp))

				fmt.Fprintf(w, "data: %s\n\n", string(data))
				w.(http.Flusher).Flush()

			case <-r.Context().Done():
				return
			}
		}
	} else {
		result, err := s.provider.Complete(r.Context(), convertCompletionRequest(req))

		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(convertChatCompletionResponse(*result))
	}
}

func convertCompletionRequest(r openai.CompletionRequest) openai.ChatCompletionRequest {
	result := openai.ChatCompletionRequest{
		Model:  r.Model,
		Stream: r.Stream,

		MaxTokens:   r.MaxTokens,
		Temperature: r.Temperature,
		TopP:        r.TopP,
		N:           r.N,

		Stop:             r.Stop,
		PresencePenalty:  r.PresencePenalty,
		FrequencyPenalty: r.FrequencyPenalty,
	}

	if prompt, ok := r.Prompt.(string); ok {
		result.Messages = []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		}
	}

	if prompts, ok := r.Prompt.([]string); ok {
		result.Messages = []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: strings.Join(prompts, "\n"),
			},
		}
	}

	return result
}

func convertChatCompletionResponse(r openai.ChatCompletionResponse) openai.CompletionResponse {
	result := openai.CompletionResponse{
		ID: r.ID,

		Object:  "text_completion",
		Created: r.Created,

		Model:   r.Model,
		Choices: []openai.CompletionChoice{},
	}

	for _, c := range r.Choices {
		result.Choices = append(result.Choices, openai.CompletionChoice{
			Index: c.Index,
			Text:  c.Message.Content,

			FinishReason: convertFinishReason(c.FinishReason),
		})
	}

	return result
}

func convertChatCompletionStreamResponse(r openai.ChatCompletionStreamResponse) openai.CompletionResponse {
	result := openai.CompletionResponse{
		ID: r.ID,

		Object:  "text_completion",
		Created: r.Created,

		Model:   r.Model,
		Choices: []openai.CompletionChoice{},
	}

	for _, c := range r.Choices {
		result.Choices = append(result.Choices, openai.CompletionChoice{
			Index: c.Index,
			Text:  c.Delta.Content,

			FinishReason: convertFinishReason(c.FinishReason),
		})
	}

	return result
}

func convertFinishReason(reason openai.FinishReason) string {
	if reason == openai.FinishReasonLength {
		return "length"
	}

	return "stop"
}
