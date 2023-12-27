package server

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/sashabaranov/go-openai"
)

func (s *Server) handleChatCompletions(w http.ResponseWriter, r *http.Request) {
	var req openai.ChatCompletionRequest

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
			done <- s.provider.CompleteStream(r.Context(), req, stream)
		}()

		for {
			select {
			case resp := <-stream:
				data, _ := json.Marshal(resp)

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
		result, err := s.provider.Complete(r.Context(), req)

		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	}
}
