package cohere

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"unicode"

	"github.com/adrianliechti/llama/pkg/provider"
)

var _ provider.Completer = (*Completer)(nil)

type Completer struct {
	*Config
}

func NewCompleter(options ...Option) (*Completer, error) {
	cfg := &Config{
		url: "https://api.cohere.com",

		client: http.DefaultClient,
	}

	for _, option := range options {
		option(cfg)
	}

	return &Completer{
		Config: cfg,
	}, nil
}

func (c *Completer) Complete(ctx context.Context, messages []provider.Message, options *provider.CompleteOptions) (*provider.Completion, error) {
	if options == nil {
		options = new(provider.CompleteOptions)
	}

	url, _ := url.JoinPath(c.url, "/v1/chat")
	body, err := convertChatRequest(c.model, messages, options)

	if err != nil {
		return nil, err
	}

	if options.Stream == nil {
		req, _ := http.NewRequestWithContext(ctx, "POST", url, jsonReader(body))
		req.Header.Set("Authorization", "Bearer "+c.token)
		req.Header.Set("Content-Type", "application/json")

		resp, err := c.client.Do(req)

		if err != nil {
			return nil, err
		}

		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, convertError(resp)
		}

		var response ChatResponse

		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			return nil, err
		}

		return &provider.Completion{
			ID:     response.ID,
			Reason: provider.CompletionReasonStop,

			Message: provider.Message{
				Role:    provider.MessageRoleAssistant,
				Content: response.Text,
			},
		}, nil
	} else {
		defer close(options.Stream)

		req, _ := http.NewRequestWithContext(ctx, "POST", url, jsonReader(body))
		req.Header.Set("Authorization", "Bearer "+c.token)
		req.Header.Set("Content-Type", "application/json")

		resp, err := c.client.Do(req)

		if err != nil {
			return nil, err
		}

		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, convertError(resp)
		}

		reader := bufio.NewReader(resp.Body)

		var resultID string

		var resultText strings.Builder
		var resultRole provider.MessageRole
		var resultReason provider.CompletionReason

		for i := 0; ; i++ {
			data, err := reader.ReadBytes('\n')

			if err != nil {
				if errors.Is(err, io.EOF) {
					break
				}

				return nil, err
			}

			data = bytes.TrimSpace(data)

			if len(data) == 0 {
				continue
			}

			var event ChatEvent

			if err := json.Unmarshal([]byte(data), &event); err != nil {
				return nil, err
			}

			if event.ID != "" {
				resultID = event.ID
			}

			var content = event.Text

			if i == 0 {
				content = strings.TrimLeftFunc(content, unicode.IsSpace)
			}

			resultText.WriteString(content)

			resultRole = provider.MessageRoleAssistant
			resultReason = toCompletionReason(event.FinishReason)

			options.Stream <- provider.Completion{
				ID:     resultID,
				Reason: resultReason,

				Message: provider.Message{
					Role:    resultRole,
					Content: content,
				},
			}
		}

		return &provider.Completion{
			ID:     resultID,
			Reason: resultReason,

			Message: provider.Message{
				Role:    resultRole,
				Content: resultText.String(),
			},
		}, nil
	}
}

func convertMessageRole(role provider.MessageRole) MessageRole {
	switch role {
	case provider.MessageRoleSystem:
		return MessageRoleSystem

	case provider.MessageRoleUser:
		return MessageRoleUser

	case provider.MessageRoleAssistant:
		return MessageRoleAssistant

	case provider.MessageRoleFunction:
		return MessageRoleTool
	}

	return ""
}

func toMessageRole(role MessageRole) provider.MessageRole {
	switch role {
	case MessageRoleSystem:
		return provider.MessageRoleSystem

	case MessageRoleUser:
		return provider.MessageRoleUser

	case MessageRoleAssistant:
		return provider.MessageRoleAssistant

	case MessageRoleTool:
		return provider.MessageRoleFunction
	}

	return ""
}

func toCompletionReason(reason FinishReason) provider.CompletionReason {
	switch reason {
	case FinishReasonComplete:
		return provider.CompletionReasonStop
	}

	return ""
}

func convertChatRequest(model string, messages []provider.Message, options *provider.CompleteOptions) (*ChatRequest, error) {
	if options == nil {
		options = new(provider.CompleteOptions)
	}

	stream := options.Stream != nil

	message := messages[len(messages)-1]

	req := &ChatRequest{
		Stream: stream,

		Model:   model,
		Message: message.Content,
	}

	for _, m := range messages[:len(messages)-1] {
		message := Message{
			Role:    convertMessageRole(m.Role),
			Message: m.Content,
		}

		req.History = append(req.History, message)
	}

	return req, nil
}

type MessageRole string

var (
	MessageRoleSystem    MessageRole = "SYSTEM"
	MessageRoleUser      MessageRole = "USER"
	MessageRoleAssistant MessageRole = "CHATBOT"
	MessageRoleTool      MessageRole = "TOOL"
)

type FinishReason string

var (
	FinishReasonComplete FinishReason = "COMPLETE"
)

type Message struct {
	Role MessageRole `json:"role,omitempty"`

	Message string `json:"message"`
}

type ChatRequest struct {
	Model string `json:"model"`

	Stream bool `json:"stream,omitempty"`

	Message string `json:"message"`

	History []Message `json:"chat_history"`
}

type ChatResponse struct {
	ID string `json:"generation_id"`

	Text string `json:"text"`

	FinishReason FinishReason `json:"finish_reason"`
}

type ChatEvent struct {
	ID string `json:"generation_id"`

	Type     string `json:"event_type"`
	Finished bool   `json:"is_finished"`

	Text string `json:"text"`

	FinishReason FinishReason `json:"finish_reason"`
}
