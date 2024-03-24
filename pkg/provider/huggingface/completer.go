package huggingface

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
	"github.com/google/uuid"
)

var _ provider.Completer = (*Completer)(nil)

type Completer struct {
	*Config
}

func NewCompleter(url string, options ...Option) (*Completer, error) {
	if url == "" {
		return nil, errors.New("invalid url")
	}

	url = strings.TrimRight(url, "/")
	url = strings.TrimSuffix(url, "/v1")

	cfg := &Config{
		url: url,

		token: "-",
		model: "tgi",

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

	id := uuid.NewString()

	url, _ := url.JoinPath(c.url, "/v1/chat/completions")
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
			return nil, errors.New("unable to complete")
		}

		var completion ChatCompletion

		if err := json.NewDecoder(resp.Body).Decode(&completion); err != nil {
			return nil, err
		}

		result := provider.Completion{
			ID:     id,
			Reason: provider.CompletionReasonStop,

			Message: provider.Message{
				Role:    provider.MessageRole(completion.Choices[0].Message.Role),
				Content: completion.Choices[0].Message.Content,
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

		var resultText strings.Builder
		var resultRole provider.MessageRole
		var resultReason provider.CompletionReason

		for i := 0; ; i++ {
			data, err := reader.ReadBytes('\n')

			if errors.Is(err, io.EOF) {
				break
			}

			if err != nil {
				return nil, err
			}

			data = bytes.TrimSpace(data)

			if bytes.HasPrefix(data, []byte("event:")) {
				continue
			}

			data = bytes.TrimPrefix(data, []byte("data:"))

			data = bytes.TrimSpace(data)

			if len(data) == 0 {
				continue
			}

			var completion ChatCompletion

			if err := json.Unmarshal([]byte(data), &completion); err != nil {
				return nil, err
			}

			var content = completion.Choices[0].Delta.Content
			content = strings.ReplaceAll(content, "<s>", "")
			content = strings.ReplaceAll(content, "</s>", "")

			if i == 0 {
				content = strings.TrimLeftFunc(content, unicode.IsSpace)
			}

			resultText.WriteString(content)

			resultRole = provider.MessageRoleAssistant
			resultReason = toCompletionReason(completion)

			options.Stream <- provider.Completion{
				ID:     id,
				Reason: resultReason,

				Message: provider.Message{
					Role:    resultRole,
					Content: content,
				},
			}
		}

		result := provider.Completion{
			ID:     id,
			Reason: resultReason,

			Message: provider.Message{
				Role:    resultRole,
				Content: resultText.String(),
			},
		}

		return &result, nil
	}
}

func convertChatRequest(model string, messages []provider.Message, options *provider.CompleteOptions) (*ChatCompletionRequest, error) {
	if options == nil {
		options = new(provider.CompleteOptions)
	}

	stream := options.Stream != nil

	req := &ChatCompletionRequest{
		Model: model,

		Stop:   options.Stop,
		Stream: stream,

		MaxTokens:   options.MaxTokens,
		Temperature: options.Temperature,
	}

	for _, m := range messages {
		message := ChatCompletionMessage{
			Role:    MessageRole(m.Role),
			Content: m.Content,
		}

		req.Messages = append(req.Messages, message)
	}

	return req, nil
}

func toCompletionReason(completion ChatCompletion) provider.CompletionReason {
	if len(completion.Choices) == 0 {
		return ""
	}

	var choice = completion.Choices[0]

	if choice.FinishReason == nil {
		return ""
	}

	switch *choice.FinishReason {
	case CompletionReasonEOS:
		return provider.CompletionReasonStop

	case CompletionReasonStop:
		return provider.CompletionReasonStop

	case CompletionReasonLength:
		return provider.CompletionReasonLength
	}

	return ""
}

type MessageRole string

var (
	MessageRoleSystem    MessageRole = "system"
	MessageRoleUser      MessageRole = "user"
	MessageRoleAssistant MessageRole = "assistant"
)

type CompletionReason string

var (
	CompletionReasonStop   CompletionReason = "stop"
	CompletionReasonLength CompletionReason = "length"
	CompletionReasonEOS    CompletionReason = "eos_token"
)

type ChatCompletionRequest struct {
	Model string `json:"model"`

	Messages []ChatCompletionMessage `json:"messages"`

	Stop   []string `json:"stop,omitempty"`
	Stream bool     `json:"stream,omitempty"`

	MaxTokens   *int     `json:"max_tokens,omitempty"`
	Temperature *float32 `json:"temperature,omitempty"`
}

type ChatCompletion struct {
	Object string `json:"object"`

	ID string `json:"id"`

	Model   string `json:"model"`
	Created int64  `json:"created"`

	Choices []ChatCompletionChoice `json:"choices"`
}

type ChatCompletionChoice struct {
	Index int `json:"index"`

	Delta   *ChatCompletionMessage `json:"delta,omitempty"`
	Message *ChatCompletionMessage `json:"message,omitempty"`

	FinishReason *CompletionReason `json:"finish_reason"`
}

type ChatCompletionMessage struct {
	Role MessageRole `json:"role,omitempty"`

	Content string `json:"content"`
}

type ChatCompletionContent struct {
	Type string `json:"type,omitempty"`
	Text string `json:"text,omitempty"`
}
