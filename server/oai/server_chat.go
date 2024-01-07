package oai

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"mime"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/adrianliechti/llama/pkg/provider"

	"github.com/google/uuid"
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

	messages, err := toMessages(req.Messages)

	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

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

func toMessages(s []ChatCompletionMessage) ([]provider.Message, error) {
	result := make([]provider.Message, 0)

	for _, m := range s {
		content := m.Content
		files := make([]provider.File, 0)

		for _, c := range m.Contents {
			if c.Type == "text" {
				content = c.Text
			}

			if c.Type == "image_url" && c.ImageURL != nil {
				file, err := toFile(*&c.ImageURL.URL)

				if err != nil {
					return nil, err
				}

				files = append(files, *file)
			}
		}

		result = append(result, provider.Message{
			Role:    toMessageRole(m.Role),
			Content: content,

			Files: files,

			Function:      m.ToolCallID,
			FunctionCalls: toFuncionCalls(m.ToolCalls),
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

	case MessageRoleTool:
		return provider.MessageRoleFunction

	default:
		return ""
	}
}

func toFile(url string) (*provider.File, error) {
	if strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://") {
		resp, err := http.Get(url)

		if err != nil {
			return nil, err
		}

		defer resp.Body.Close()

		data, err := io.ReadAll(resp.Body)

		if err != nil {
			return nil, err
		}

		file := provider.File{
			Content: bytes.NewReader(data),
		}

		if ext, _ := mime.ExtensionsByType(resp.Header.Get("Content-Type")); len(ext) > 0 {
			file.Name = uuid.New().String() + ext[0]
		}

		return &file, nil
	}

	if strings.HasPrefix(url, "data:") {
		re := regexp.MustCompile(`data:([a-zA-Z]+\/[a-zA-Z0-9.+_-]+);base64,\s*(.+)`)

		match := re.FindStringSubmatch(url)

		if len(match) != 3 {
			return nil, fmt.Errorf("invalid data url")
		}

		data, err := base64.StdEncoding.DecodeString(match[2])

		if err != nil {
			return nil, fmt.Errorf("invalid data encoding")
		}

		file := provider.File{
			Content: bytes.NewReader(data),
		}

		if ext, _ := mime.ExtensionsByType(match[1]); len(ext) > 0 {
			file.Name = uuid.New().String() + ext[0]
		}

		return &file, nil
	}

	return nil, fmt.Errorf("invalid url")
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
