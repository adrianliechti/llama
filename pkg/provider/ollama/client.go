package ollama

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

var (
	_ provider.Embedder  = (*Client)(nil)
	_ provider.Completer = (*Client)(nil)
)

type Client struct {
	url string

	model  string
	system string

	client *http.Client
}

type Option func(*Client)

func New(options ...Option) (*Client, error) {
	c := &Client{
		url: "http://localhost:11434",

		client: http.DefaultClient,
	}

	for _, option := range options {
		option(c)
	}

	if c.url == "" {
		return nil, errors.New("invalid url")
	}

	return c, nil
}

func WithClient(client *http.Client) Option {
	return func(c *Client) {
		c.client = client
	}
}

func WithURL(url string) Option {
	return func(c *Client) {
		c.url = url
	}
}

func WithModel(model string) Option {
	return func(c *Client) {
		c.model = model
	}
}

func WithSystem(system string) Option {
	return func(c *Client) {
		c.system = system
	}
}

func (c *Client) Embed(ctx context.Context, content string) ([]float32, error) {
	body := &EmbeddingRequest{
		Model:  c.model,
		Prompt: strings.TrimSpace(content),
	}

	u, _ := url.JoinPath(c.url, "/api/embeddings")
	resp, err := c.client.Post(u, "application/json", jsonReader(body))

	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("unable to embed")
	}

	defer resp.Body.Close()

	var result EmbeddingResponse

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return toFloat32s(result.Embedding), nil
}

func (c *Client) Complete(ctx context.Context, messages []provider.Message, options *provider.CompleteOptions) (*provider.Completion, error) {
	if options == nil {
		options = new(provider.CompleteOptions)
	}

	id := uuid.NewString()

	url, _ := url.JoinPath(c.url, "/api/chat")
	body, err := c.convertChatRequest(messages, options)

	if err != nil {
		return nil, err
	}

	if options.Stream == nil {
		resp, err := c.client.Post(url, "application/json", jsonReader(body))

		if err != nil {
			return nil, err
		}

		if resp.StatusCode != http.StatusOK {
			return nil, errors.New("unable to complete")
		}

		defer resp.Body.Close()

		var chat ChatResponse

		if err := json.NewDecoder(resp.Body).Decode(&chat); err != nil {
			return nil, err
		}

		result := provider.Completion{
			ID: id,

			Reason: provider.CompletionReasonStop,

			Message: provider.Message{
				Role:    provider.MessageRole(chat.Message.Role),
				Content: chat.Message.Content,
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

			resultText.WriteString(content)

			resultRole = provider.MessageRoleAssistant
			resultReason = toCompletionReason(chat)

			options.Stream <- provider.Completion{
				ID: id,

				Reason: resultReason,

				Message: provider.Message{
					Role:    resultRole,
					Content: content,
				},
			}
		}

		result := provider.Completion{
			Reason: resultReason,

			Message: provider.Message{
				Role:    resultRole,
				Content: resultText.String(),
			},
		}

		return &result, nil
	}
}

func (c *Client) convertChatRequest(messages []provider.Message, options *provider.CompleteOptions) (*ChatRequest, error) {
	if options == nil {
		options = new(provider.CompleteOptions)
	}

	if c.system != "" && len(messages) > 0 && messages[0].Role != provider.MessageRoleSystem {
		message := provider.Message{
			Role:    provider.MessageRoleSystem,
			Content: c.system,
		}

		messages = append([]provider.Message{message}, messages...)
	}

	stream := options.Stream != nil

	req := &ChatRequest{
		Model:  c.model,
		Stream: &stream,

		Options: map[string]any{},
	}

	if options.Stop != nil {
		req.Options["stop"] = options.Stop
	}

	if options.Format == provider.CompletionFormatJSON {
		req.Format = "json"
	}

	for _, m := range messages {
		message := Message{
			Role:    MessageRole(m.Role),
			Content: m.Content,
		}

		for _, f := range m.Files {
			data, err := io.ReadAll(f.Content)

			if err != nil {
				return nil, err
			}

			message.Images = append(message.Images, data)
		}

		req.Messages = append(req.Messages, message)
	}

	return req, nil
}

func jsonReader(v any) io.Reader {
	b := new(bytes.Buffer)

	enc := json.NewEncoder(b)
	enc.SetEscapeHTML(false)

	enc.Encode(v)
	return b
}

func toFloat32s(v []float64) []float32 {
	result := make([]float32, len(v))

	for i, x := range v {
		result[i] = float32(x)
	}

	return result
}

func toCompletionReason(chat ChatResponse) provider.CompletionReason {
	if chat.Done {
		return provider.CompletionReasonStop
	}

	return ""
}
