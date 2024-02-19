package llama

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
	"github.com/adrianliechti/llama/pkg/provider/llama/grammar"
	"github.com/adrianliechti/llama/pkg/provider/llama/prompt"

	"github.com/google/uuid"
)

var (
	_ provider.Embedder  = (*Client)(nil)
	_ provider.Completer = (*Client)(nil)
)

type Client struct {
	url string

	client *http.Client

	system   string
	template prompt.Template
}

type Option func(*Client)

type Template = prompt.Template

var (
	TemplateNone   = prompt.None
	TemplateSimple = prompt.Simple

	TemplateChatML = prompt.ChatML
	TemplateToRA   = prompt.ToRA

	TemplateLlama   = prompt.Llama
	TemplateMistral = prompt.Mistral

	TemplateGorilla    = prompt.Gorilla
	TemplateNexusRaven = prompt.NexusRaven
	TemplateLlamaGuard = prompt.LlamaGuard
)

func New(url string, options ...Option) (*Client, error) {
	if url == "" {
		return nil, errors.New("invalid url")
	}

	p := &Client{
		url: url,

		client: http.DefaultClient,

		system:   "",
		template: prompt.Llama,
	}

	for _, option := range options {
		option(p)
	}

	return p, nil
}

func WithClient(client *http.Client) Option {
	return func(c *Client) {
		c.client = client
	}
}

func WithSystem(system string) Option {
	return func(c *Client) {
		c.system = system
	}
}

func WithTemplate(template Template) Option {
	return func(c *Client) {
		c.template = template
	}
}

func (c *Client) Embed(ctx context.Context, content string) ([]float32, error) {
	body := &EmbeddingRequest{
		Content: strings.TrimSpace(content),
	}

	u, _ := url.JoinPath(c.url, "/embedding")
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

	return result.Embedding, nil
}

func (c *Client) Complete(ctx context.Context, messages []provider.Message, options *provider.CompleteOptions) (*provider.Completion, error) {
	if options == nil {
		options = &provider.CompleteOptions{}
	}

	id := uuid.NewString()

	url, _ := url.JoinPath(c.url, "/completion")
	body, err := c.convertCompletionRequest(messages, options)

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

		var completion CompletionResponse

		if err := json.NewDecoder(resp.Body).Decode(&completion); err != nil {
			return nil, err
		}

		content := strings.TrimSpace(completion.Content)

		var resultRole = provider.MessageRoleAssistant
		var resultReason = toCompletionReason(completion)

		result := provider.Completion{
			ID: id,

			Reason: resultReason,

			Message: provider.Message{
				Role:    resultRole,
				Content: content,
			},
		}

		return &result, nil
	} else {
		defer close(options.Stream)

		req, _ := http.NewRequestWithContext(ctx, "POST", url, jsonReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "text/event-stream")

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

			// if bytes.HasPrefix(data, errorPrefix) {
			// }

			data = bytes.TrimPrefix(data, []byte("data: "))

			if len(data) == 0 {
				continue
			}

			var completion CompletionResponse

			if err := json.Unmarshal([]byte(data), &completion); err != nil {
				return nil, err
			}

			var content = completion.Content

			if i == 0 {
				content = strings.TrimLeftFunc(content, unicode.IsSpace)
			}

			resultText.WriteString(content)

			resultRole = provider.MessageRoleAssistant
			resultReason = toCompletionReason(completion)

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
			ID: id,

			Reason: resultReason,

			Message: provider.Message{
				Role:    resultRole,
				Content: resultText.String(),
			},
		}

		return &result, nil
	}
}

func (c *Client) convertCompletionRequest(messages []provider.Message, options *provider.CompleteOptions) (*CompletionRequest, error) {
	if options == nil {
		options = &provider.CompleteOptions{}
	}

	prompt, err := c.template.Prompt(c.system, messages, toTemplateOptions(options))

	if err != nil {
		return nil, err
	}

	req := &CompletionRequest{
		Prompt: prompt,

		Stream: options.Stream != nil,

		Temperature: options.Temperature,
		TopP:        options.TopP,
		MinP:        options.MinP,

		Stop: c.template.Stop(),

		CachePrompt: true,
	}

	for _, m := range messages {
		for i, f := range m.Files {
			data, err := io.ReadAll(f.Content)

			if err != nil {
				return nil, err
			}
			_ = i

			req.Images = append(req.Images, CompletionImage{
				ID:   i + 1,
				Data: data,
			})
		}
	}

	if options.Stop != nil {
		req.Stop = options.Stop
	}

	if options.Format == provider.CompletionFormatJSON {
		req.Grammar = grammar.JSON
	}

	return req, nil
}

func toCompletionReason(resp CompletionResponse) provider.CompletionReason {
	if resp.Truncated {
		return provider.CompletionReasonLength
	}

	if resp.Stop {
		return provider.CompletionReasonStop
	}

	return ""
}

func toTemplateOptions(options *provider.CompleteOptions) *prompt.TemplateOptions {
	if options == nil {
		return nil
	}

	return &prompt.TemplateOptions{
		Functions: options.Functions,
	}
}

func jsonReader(v any) io.Reader {
	b := new(bytes.Buffer)

	enc := json.NewEncoder(b)
	enc.SetEscapeHTML(false)

	enc.Encode(v)
	return b
}
