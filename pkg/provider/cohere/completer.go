package cohere

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/adrianliechti/llama/pkg/provider"
)

var _ provider.Completer = (*Completer)(nil)

type Completer struct {
	*Config
}

func NewCompleter(model string, options ...Option) (*Completer, error) {
	cfg := &Config{
		client: http.DefaultClient,

		url:   "https://api.cohere.com",
		model: model,
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

		result := &provider.Completion{
			Message: provider.Message{
				Role: provider.MessageRoleAssistant,
			},
		}

		for i := 0; ; i++ {
			data, err := reader.ReadString('\n')

			if err != nil {
				if errors.Is(err, io.EOF) {
					break
				}

				return nil, err
			}

			data = strings.TrimSpace(data)

			if len(data) == 0 {
				continue
			}

			var event ChatEvent

			if err := json.Unmarshal([]byte(data), &event); err != nil {
				return nil, err
			}

			if event.ID != "" {
				result.ID = event.ID
			}

			result.Reason = toCompletionReason(event.FinishReason)
			result.Message.Content += event.Text

			options.Stream <- provider.Completion{
				ID:     result.ID,
				Reason: result.Reason,

				Message: provider.Message{
					Role:    result.Message.Role,
					Content: event.Text,
				},
			}
		}

		if result.Usage.OutputTokens == 0 {
			result.Usage = nil
		}

		return result, nil
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

	case provider.MessageRoleTool:
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
		return provider.MessageRoleTool
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

		MaxTokens:   options.MaxTokens,
		Temperature: options.Temperature,

		StopSequences: options.Stop,
	}

	for _, m := range messages[:len(messages)-1] {
		message := Message{
			Role:    convertMessageRole(m.Role),
			Message: m.Content,
		}

		req.History = append(req.History, message)
	}

	if options.Format == provider.CompletionFormatJSON {
		req.ResponseFormat = ResponseFormatJSON
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

type ResponseFormat string

var (
	ResponseFormatText ResponseFormat = "text"
	ResponseFormatJSON ResponseFormat = "json_object"
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
	Stream bool `json:"stream,omitempty"`

	Model   string    `json:"model"`
	Message string    `json:"message"`
	History []Message `json:"chat_history"`

	MaxTokens   *int     `json:"max_tokens,omitempty"`
	Temperature *float32 `json:"temperature,omitempty"`

	StopSequences  []string       `json:"stop_sequences,omitempty"`
	ResponseFormat ResponseFormat `json:"response_format,omitempty"`
}

type ChatResponse struct {
	ID string `json:"generation_id"`

	Text string `json:"text"`

	FinishReason FinishReason `json:"finish_reason"`

	Metadata Metadata `json:"meta"`
}

type ChatEvent struct {
	ID string `json:"generation_id"`

	Type     string `json:"event_type"`
	Finished bool   `json:"is_finished"`

	Text string `json:"text"`

	FinishReason FinishReason `json:"finish_reason"`
}

type Metadata struct {
	Usage Usage `json:"billed_units"`
}

type Usage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}
