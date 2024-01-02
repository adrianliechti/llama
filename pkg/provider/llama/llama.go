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
)

var (
	_ provider.Provider = &Provider{}
)

type Provider struct {
	url string

	client *http.Client

	system   string
	template prompt.Template
}

type Option func(*Provider)

type Template = prompt.Template

var (
	TemplateChatML     = prompt.ChatML
	TemplateLlama      = prompt.Llama
	TemplateLlamaGuard = prompt.LlamaGuard
	TemplateMistral    = prompt.Mistral
)

func New(url string, options ...Option) (*Provider, error) {
	p := &Provider{
		url: url,

		client: http.DefaultClient,

		system:   "",
		template: prompt.Llama,
	}

	for _, option := range options {
		option(p)
	}

	if p.url == "" {
		return nil, errors.New("invalid url")
	}

	return p, nil
}

func WithClient(client *http.Client) Option {
	return func(p *Provider) {
		p.client = client
	}
}

func WithSystem(system string) Option {
	return func(p *Provider) {
		p.system = system
	}
}

func WithTemplate(template Template) Option {
	return func(p *Provider) {
		p.template = template
	}
}

func (p *Provider) Embed(ctx context.Context, model, content string) ([]float32, error) {
	body := &EmbeddingRequest{
		Content: strings.TrimSpace(content),
	}

	u, _ := url.JoinPath(p.url, "/embedding")
	resp, err := p.client.Post(u, "application/json", jsonReader(body))

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

func (p *Provider) Complete(ctx context.Context, model string, messages []provider.Message, options *provider.CompleteOptions) (*provider.Completion, error) {
	if options == nil {
		options = &provider.CompleteOptions{}
	}

	url, _ := url.JoinPath(p.url, "/completion")
	body, err := p.convertCompletionRequest(messages, options)

	if err != nil {
		return nil, err
	}

	if options.Stream == nil {
		resp, err := p.client.Post(url, "application/json", jsonReader(body))

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
			Message: &provider.Message{
				Role:    resultRole,
				Content: content,
			},

			Reason: resultReason,
		}

		return &result, nil
	} else {
		defer close(options.Stream)

		req, _ := http.NewRequestWithContext(ctx, "POST", url, jsonReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "text/event-stream")

		resp, err := p.client.Do(req)

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
				Message: &provider.Message{
					Role:    resultRole,
					Content: content,
				},

				Reason: resultReason,
			}
		}

		result := provider.Completion{
			Message: &provider.Message{
				Role:    resultRole,
				Content: resultText.String(),
			},

			Reason: resultReason,
		}

		return &result, nil
	}
}

func (p *Provider) convertCompletionRequest(messages []provider.Message, options *provider.CompleteOptions) (*CompletionRequest, error) {
	if options == nil {
		options = &provider.CompleteOptions{}
	}

	prompt, err := p.template.Prompt(p.system, messages)

	if err != nil {
		return nil, err
	}

	req := &CompletionRequest{
		Prompt: prompt,

		Stream: options.Stream != nil,

		Temperature: options.Temperature,
		TopP:        options.TopP,
		MinP:        options.MinP,

		Stop: p.template.Stop(),
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

func jsonReader(v any) io.Reader {
	b := new(bytes.Buffer)

	enc := json.NewEncoder(b)
	enc.SetEscapeHTML(false)

	enc.Encode(v)
	return b
}
