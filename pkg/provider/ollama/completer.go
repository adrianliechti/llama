package ollama

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"unicode"

	"github.com/adrianliechti/llama/pkg/provider"

	"github.com/google/uuid"
)

var _ provider.Completer = (*Completer)(nil)

type Completer struct {
	*Config
}

func NewCompleter(url string, options ...Option) (*Completer, error) {
	if url == "" {
		url = "http://localhost:11434"
	}

	c := &Config{
		url:    url,
		client: http.DefaultClient,
	}

	for _, option := range options {
		option(c)
	}

	go c.ensureModel()

	return &Completer{
		Config: c,
	}, nil
}

func (c *Completer) Complete(ctx context.Context, messages []provider.Message, options *provider.CompleteOptions) (*provider.Completion, error) {
	if options == nil {
		options = new(provider.CompleteOptions)
	}

	id := uuid.NewString()

	url, _ := url.JoinPath(c.url, "/api/chat")
	body, err := convertChatRequest(c.model, messages, options)

	if err != nil {
		return nil, err
	}

	if options.Stream == nil {
		resp, err := c.client.Post(url, "application/json", jsonReader(body))

		if err != nil {
			return nil, err
		}

		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			data, _ := io.ReadAll(resp.Body)

			if len(data) == 0 {
				data = []byte("unable to complete")
			}

			return nil, errors.New(string(data))
		}

		var chat ChatResponse

		if err := json.NewDecoder(resp.Body).Decode(&chat); err != nil {
			return nil, err
		}

		role := toMessageRole(chat.Message.Role)
		content := strings.TrimSpace(chat.Message.Content)

		if role == "" {
			role = provider.MessageRoleAssistant
		}

		result := provider.Completion{
			ID:     id,
			Reason: provider.CompletionReasonStop,

			Message: provider.Message{
				Role:    role,
				Content: content,
			},
		}

		return &result, nil
	} else {
		defer close(options.Stream)

		req, _ := http.NewRequestWithContext(ctx, "POST", url, jsonReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/x-ndjson")

		resp, err := c.client.Do(req)

		if err != nil {
			return nil, err
		}

		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, errors.New("unable to complete")
		}

		reader := bufio.NewReader(resp.Body)

		result := provider.Completion{
			ID: id,

			Message: provider.Message{
				Role: provider.MessageRoleAssistant,
			},
		}

		for i := 0; ; i++ {
			data, err := reader.ReadBytes('\n')

			if errors.Is(err, io.EOF) {
				break
			}

			if err != nil {
				return nil, err
			}

			if len(data) == 0 {
				continue
			}

			var chat ChatResponse

			if err := json.Unmarshal([]byte(data), &chat); err != nil {
				return nil, err
			}

			var content = chat.Message.Content

			if i == 0 {
				content = strings.TrimLeftFunc(content, unicode.IsSpace)
			}

			role := toMessageRole(chat.Message.Role)

			if role == "" {
				role = provider.MessageRoleAssistant
			}

			result.Reason = toCompletionReason(chat)

			result.Message.Role = role
			result.Message.Content += content

			options.Stream <- provider.Completion{
				ID:     result.ID,
				Reason: result.Reason,

				Message: provider.Message{
					Role:    role,
					Content: content,
				},
			}
		}

		return &result, nil
	}
}

func convertChatRequest(model string, messages []provider.Message, options *provider.CompleteOptions) (*ChatRequest, error) {
	if options == nil {
		options = new(provider.CompleteOptions)
	}

	stream := options.Stream != nil

	req := &ChatRequest{
		Model:  model,
		Stream: &stream,

		Options: map[string]any{},
	}

	if options.Stop != nil {
		req.Options["stop"] = options.Stop
	}

	if options.Format == provider.CompletionFormatJSON {
		req.Format = "json"
	}

	if options.MaxTokens != nil {
		req.Options["num_predict"] = *options.MaxTokens
	}

	if options.Temperature != nil {
		req.Options["temperature"] = *options.Temperature
	}

	for i, m := range messages {
		message := Message{
			Role:    convertMessageRole(m.Role),
			Content: m.Content,
		}

		// HACK: only use images on last message
		if i == len(messages)-1 {
			for _, f := range m.Files {
				data, err := io.ReadAll(f.Content)

				if err != nil {
					return nil, err
				}

				message.Images = append(message.Images, data)
			}
		}

		req.Messages = append(req.Messages, message)
	}

	return req, nil
}

func convertMessageRole(r provider.MessageRole) MessageRole {
	switch r {

	case provider.MessageRoleSystem:
		return MessageRoleSystem

	case provider.MessageRoleUser:
		return MessageRoleUser

	case provider.MessageRoleAssistant:
		return MessageRoleAssistant

	// case provider.MessageRoleFunction:
	// 	return MessageRoleTool

	default:
		return ""
	}
}

func toMessageRole(role MessageRole) provider.MessageRole {
	switch role {

	case MessageRoleSystem:
		return provider.MessageRoleSystem

	case MessageRoleUser:
		return provider.MessageRoleUser

	case MessageRoleAssistant:
		return provider.MessageRoleAssistant

	// case MessageRoleTool:
	// 	return provider.MessageRoleFunction

	default:
		return ""
	}
}

func toCompletionReason(chat ChatResponse) provider.CompletionReason {
	if chat.Done {
		return provider.CompletionReasonStop
	}

	return ""
}

type MessageRole string

var (
	MessageRoleSystem    MessageRole = "system"
	MessageRoleUser      MessageRole = "user"
	MessageRoleAssistant MessageRole = "assistant"
	//MessageRoleTool      MessageRole = "tool"
)

type MessageImage []byte

type Message struct {
	Role    MessageRole `json:"role"`
	Content string      `json:"content"`

	Images []MessageImage `json:"images,omitempty"`
}

type ChatRequest struct {
	Model string `json:"model"`

	Stream *bool  `json:"stream,omitempty"`
	Format string `json:"format,omitempty"`

	Messages []Message `json:"messages"`

	Options map[string]interface{} `json:"options"`
}

type ChatResponse struct {
	Model   string  `json:"model"`
	Message Message `json:"message"`

	Done bool `json:"done"`
}
