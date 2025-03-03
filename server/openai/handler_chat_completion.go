package openai

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/adrianliechti/wingman/pkg/provider"

	"github.com/google/uuid"
)

func (h *Handler) handleChatCompletion(w http.ResponseWriter, r *http.Request) {
	var req ChatCompletionRequest

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

	tools, err := toTools(req.Tools)

	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	var stops []string

	switch v := req.Stop.(type) {
	case string:
		stops = []string{v}

	case []string:
		stops = v
	}

	options := &provider.CompleteOptions{
		Stop:  stops,
		Tools: tools,

		MaxTokens:   req.MaxTokens,
		Temperature: req.Temperature,
	}

	switch req.ReasoningEffort {
	case ReasoningEffortLow:
		options.Effort = provider.ReasoningEffortLow
	case ReasoningEffortMedium:
		options.Effort = provider.ReasoningEffortMedium
	case ReasoningEffortHigh:
		options.Effort = provider.ReasoningEffortHigh
	}

	if req.ResponseFormat != nil {
		if req.ResponseFormat.Type == ResponseFormatJSONObject || req.ResponseFormat.Type == ResponseFormatJSONSchema {
			options.Format = provider.CompletionFormatJSON
		}

		if req.ResponseFormat.JSONSchema != nil {
			options.Format = provider.CompletionFormatJSON

			options.Schema = &provider.Schema{
				Name:        req.ResponseFormat.JSONSchema.Name,
				Description: req.ResponseFormat.JSONSchema.Description,

				Strict: req.ResponseFormat.JSONSchema.Strict,
				Schema: req.ResponseFormat.JSONSchema.Schema,
			}
		}
	}

	if req.Stream {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		rc := http.NewResponseController(w)

		options.Stream = func(ctx context.Context, completion provider.Completion) error {
			result := ChatCompletion{
				Object: "chat.completion.chunk",

				ID: completion.ID,

				Model:   req.Model,
				Created: time.Now().Unix(),

				Choices: []ChatCompletionChoice{
					{
						FinishReason: oaiFinishReason(completion.Reason),

						Delta: &ChatCompletionMessage{
							Role:    oaiMessageRole(completion.Message.Role),
							Content: completion.Message.Content,

							ToolCalls:  oaiToolCalls(completion.Message.ToolCalls),
							ToolCallID: completion.Message.Tool,
						},
					},
				},
			}

			if completion.Usage != nil {
				result.Usage = &Usage{
					PromptTokens:     completion.Usage.InputTokens,
					CompletionTokens: completion.Usage.OutputTokens,
					TotalTokens:      completion.Usage.InputTokens + completion.Usage.OutputTokens,
				}
			}

			var data bytes.Buffer

			enc := json.NewEncoder(&data)
			enc.SetEscapeHTML(false)
			enc.Encode(result)

			event := strings.TrimSpace(data.String())

			if _, err := fmt.Fprintf(w, "data: %s\n\n", event); err != nil {
				return err
			}

			if err := rc.Flush(); err != nil {
				return err
			}

			return nil
		}

		if _, err := completer.Complete(r.Context(), messages, options); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}

		fmt.Fprintf(w, "data: [DONE]\n\n")
		rc.Flush()
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
					FinishReason: oaiFinishReason(completion.Reason),

					Message: &ChatCompletionMessage{
						Role:    oaiMessageRole(completion.Message.Role),
						Content: completion.Message.Content,

						ToolCalls:  oaiToolCalls(completion.Message.ToolCalls),
						ToolCallID: completion.Message.Tool,
					},
				},
			},
		}

		if completion.Usage != nil {
			result.Usage = &Usage{
				PromptTokens:     completion.Usage.InputTokens,
				CompletionTokens: completion.Usage.OutputTokens,
				TotalTokens:      completion.Usage.InputTokens + completion.Usage.OutputTokens,
			}
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
				if len(content) > 0 {
					content += "\n\n"
				}

				content += c.Text
			}

			if c.Type == MessageContentTypeFileURL && c.FileURL != nil {
				file, err := toFile(c.FileURL.URL)

				if err != nil {
					return nil, err
				}

				files = append(files, *file)
			}

			if c.Type == MessageContentTypeImageURL && c.ImageURL != nil {
				file, err := toFile(c.ImageURL.URL)

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

			Tool:      m.ToolCallID,
			ToolCalls: toToolCalls(m.ToolCalls),
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
		return provider.MessageRoleTool

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
			Content:     bytes.NewReader(data),
			ContentType: resp.Header.Get("Content-Type"),
		}

		if ext, _ := mime.ExtensionsByType(file.ContentType); len(ext) > 0 {
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
			Content:     bytes.NewReader(data),
			ContentType: match[1],
		}

		if ext, _ := mime.ExtensionsByType(file.ContentType); len(ext) > 0 {
			file.Name = uuid.New().String() + ext[0]
		}

		return &file, nil
	}

	return nil, fmt.Errorf("invalid url")
}

func toTools(tools []Tool) ([]provider.Tool, error) {
	var result []provider.Tool

	for _, t := range tools {
		if t.Type == ToolTypeFunction && t.ToolFunction != nil {
			function := provider.Tool{
				Name:        t.ToolFunction.Name,
				Description: t.ToolFunction.Description,

				Parameters: t.ToolFunction.Parameters,
			}

			result = append(result, function)
		}
	}

	return result, nil
}

func toToolCalls(calls []ToolCall) []provider.ToolCall {
	var result []provider.ToolCall

	for _, c := range calls {
		if c.Type == ToolTypeFunction && c.Function != nil {
			call := provider.ToolCall{
				ID: c.ID,

				Name:      c.Function.Name,
				Arguments: c.Function.Arguments,
			}

			result = append(result, call)
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

	case provider.MessageRoleTool:
		return MessageRoleTool

	default:
		return ""
	}
}

func oaiFinishReason(val provider.CompletionReason) *FinishReason {
	switch val {
	case provider.CompletionReasonStop:
		return &FinishReasonStop

	case provider.CompletionReasonLength:
		return &FinishReasonLength

	case provider.CompletionReasonTool:
		return &FinishReasonToolCalls

	case provider.CompletionReasonFilter:
		return &FinishReasonContentFilter

	default:
		return nil
	}
}

func oaiToolCalls(calls []provider.ToolCall) []ToolCall {
	result := make([]ToolCall, 0)

	for i, c := range calls {
		result = append(result, ToolCall{
			Index: i,

			ID: c.ID,

			Type: ToolTypeFunction,

			Function: &FunctionCall{
				Name:      c.Name,
				Arguments: c.Arguments,
			},
		})
	}

	return result
}
