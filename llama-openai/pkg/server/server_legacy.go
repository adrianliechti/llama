package server

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/sashabaranov/go-openai"
)

func (s *Server) handleCompletions(w http.ResponseWriter, r *http.Request) {
	var legacyReq openai.CompletionRequest

	if err := json.NewDecoder(r.Body).Decode(&legacyReq); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	req := convertChatRequest(legacyReq)

	if req.Stream {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		done := make(chan error)
		stream := make(chan openai.ChatCompletionStreamResponse)

		// defer func() {
		// 	close(done)
		// 	close(stream)
		// }()

		go func() {
			done <- s.provider.ChatStream(r.Context(), req, stream)
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
		result, err := s.provider.Chat(r.Context(), req)

		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(convertChatCompletionResponse(*result))
	}
}

func convertChatRequest(r openai.CompletionRequest) openai.ChatCompletionRequest {
	result := openai.ChatCompletionRequest{
		Model: r.Model,

		MaxTokens:   r.MaxTokens,
		Temperature: r.Temperature,
		TopP:        r.TopP,
		N:           r.N,

		Stream: r.Stream,

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
		for _, prompt := range prompts {
			result.Messages = []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			}
		}
	}

	return result
}

func convertChatCompletionResponse(r openai.ChatCompletionResponse) openai.CompletionResponse {
	result := openai.CompletionResponse{
		ID:      r.ID,
		Object:  r.Object,
		Created: r.Created,
		Model:   r.Model,

		Choices: []openai.CompletionChoice{},
	}

	for _, c := range r.Choices {
		result.Choices = append(result.Choices, openai.CompletionChoice{
			Index: c.Index,
			Text:  c.Message.Content,

			FinishReason: string(c.FinishReason),
		})
	}

	return result
}

func convertChatCompletionStreamResponse(r openai.ChatCompletionStreamResponse) openai.CompletionResponse {
	result := openai.CompletionResponse{
		ID:      r.ID,
		Object:  r.Object,
		Created: r.Created,
		Model:   r.Model,

		Choices: []openai.CompletionChoice{},
	}

	for _, c := range r.Choices {
		result.Choices = append(result.Choices, openai.CompletionChoice{
			Index: c.Index,
			Text:  c.Delta.Content,

			FinishReason: string(c.FinishReason),
		})
	}

	return result
}
