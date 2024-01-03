package oai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/adrianliechti/llama/pkg/provider"
)

func (s *Server) handleChatCompletions(w http.ResponseWriter, r *http.Request) {
	var req ChatCompletionRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	completer, err := s.Completer(req.Model)

	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	messages := toMessages(req.Messages)

	options := &provider.CompleteOptions{
		Temperature: req.Temperature,
		TopP:        req.TopP,

		Functions: toFunctions(req.Tools),
	}

	if req.Format != nil {
		if req.Format.Type == ResponseFormatJSON {
			options.Format = provider.CompletionFormatJSON
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
			options.Stream = stream

			_, err := completer.Complete(r.Context(), messages, options)
			done <- err
		}()

		for completion := range stream {
			result := ChatCompletion{
				Object: "chat.completion.chunk",

				ID: completion.ID,

				Model:   req.Model,
				Created: time.Now().Unix(),

				Choices: []ChatCompletionChoice{
					{
						FinishReason: oaiCompletionReason(completion.Reason),

						Delta: &ChatCompletionMessage{
							//Role:    fromMessageRole(completion.Role),
							Content: completion.Message.Content,

							ToolCalls:  oaiToolCalls(completion.Message.FunctionCalls),
							ToolCallID: completion.Message.Function,
						},
					},
				},
			}

			var data bytes.Buffer

			enc := json.NewEncoder(&data)
			enc.SetEscapeHTML(false)
			enc.Encode(result)

			fmt.Fprintf(w, "data: %s\n\n", data.String())
			w.(http.Flusher).Flush()
		}

		// fmt.Fprintf(w, "data: [DONE]\n\n")
		// w.(http.Flusher).Flush()

		if err := <-done; err != nil {
			slog.Error("error in chat completion", "error", err)
		}

	} else {
		completion, err := completer.Complete(r.Context(), messages, options)

		if err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}

		result := ChatCompletion{
			Object: "chat.completion",

			ID: completion.ID,

			Model:   req.Model,
			Created: time.Now().Unix(),

			Choices: []ChatCompletionChoice{
				{
					FinishReason: oaiCompletionReason(completion.Reason),

					Message: &ChatCompletionMessage{
						Role:    oaiMessageRole(completion.Message.Role),
						Content: completion.Message.Content,

						ToolCalls:  oaiToolCalls(completion.Message.FunctionCalls),
						ToolCallID: completion.Message.Function,
					},
				},
			},
		}

		writeJson(w, result)
	}
}

func toMessages(s []ChatCompletionMessage) []provider.Message {
	result := make([]provider.Message, 0)

	for _, m := range s {
		result = append(result, provider.Message{
			Role:    toMessageRole(m.Role),
			Content: m.Content,

			Function:      m.ToolCallID,
			FunctionCalls: toFuncionCalls(m.ToolCalls),
		})
	}

	return result
}

func toMessageRole(r MessageRole) provider.MessageRole {
	switch r {
	case MessageRoleSystem:
		return provider.MessageRoleSystem

	case MessageRoleUser:
		return provider.MessageRoleUser

	case MessageRoleAssistant:
		return provider.MessageRoleAssistant

	case MessageRoleTool:
		return provider.MessageRoleFunction

	default:
		return ""
	}
}

func toFunctions(s []Tool) []provider.Function {
	var result []provider.Function

	for _, t := range s {
		if t.Type == ToolTypeFunction && t.ToolFunction != nil {
			function := provider.Function{
				Name:       t.ToolFunction.Name,
				Parameters: t.ToolFunction.Parameters,

				Description: t.ToolFunction.Description,
			}

			result = append(result, function)
		}
	}

	return result
}

func toFuncionCalls(calls []ToolCall) []provider.FunctionCall {
	var result []provider.FunctionCall

	for _, c := range calls {
		if c.Type == ToolTypeFunction && c.Function != nil {
			result = append(result, provider.FunctionCall{
				ID: c.ID,

				Name:      c.Function.Name,
				Arguments: c.Function.Arguments,
			})
		}
	}

	return result
}

func oaiMessageRole(r provider.MessageRole) MessageRole {
	switch r {
	case provider.MessageRoleSystem:
		return MessageRoleSystem

	case provider.MessageRoleUser:
		return MessageRoleUser

	case provider.MessageRoleAssistant:
		return MessageRoleAssistant

	case provider.MessageRoleFunction:
		return MessageRoleTool

	default:
		return ""
	}
}

func oaiCompletionReason(val provider.CompletionReason) *CompletionReason {
	switch val {
	case provider.CompletionReasonStop:
		return &CompletionReasonStop

	case provider.CompletionReasonLength:
		return &CompletionReasonLength

	case provider.CompletionReasonFunction:
		return &CompletionReasonToolCalls

	default:
		return nil
	}
}

func oaiToolCalls(calls []provider.FunctionCall) []ToolCall {
	result := make([]ToolCall, 0)

	for _, c := range calls {
		result = append(result, ToolCall{
			ID:   c.ID,
			Type: ToolTypeFunction,

			Function: &FunctionCall{
				Name:      c.Name,
				Arguments: c.Arguments,
			},
		})
	}

	return result
}
