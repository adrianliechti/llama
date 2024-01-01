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
)

var (
	_ provider.Provider = &Provider{}
)

type Provider struct {
	url string

	client *http.Client

	system   string
	template PromptTemplate
}

type Option func(*Provider)

var (
	headerData = []byte("data: ")
	//errorPrefix = []byte(`data: {"error":`)
)

func New(url string, options ...Option) (*Provider, error) {
	p := &Provider{
		url: url,

		client: http.DefaultClient,

		system:   "",
		template: &PromptLlama{},
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

func WithPromptTemplate(template PromptTemplate) Option {
	return func(p *Provider) {
		p.template = template
	}
}

func (p *Provider) Embed(ctx context.Context, model, content string) ([]float32, error) {
	body := &embeddingRequest{
		Content: strings.TrimSpace(content),
	}

	u, _ := url.JoinPath(p.url, "/embedding")
	resp, err := p.client.Post(u, "application/json", jsonReader(body))

	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("unable to complete chat")
	}

	defer resp.Body.Close()

	var result embeddingResponse

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

		var completion completionResponse

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

			data = bytes.TrimPrefix(data, headerData)

			if len(data) == 0 {
				continue
			}

			var completion completionResponse

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

func (p *Provider) convertCompletionRequest(messages []provider.Message, options *provider.CompleteOptions) (*completionRequest, error) {
	if options == nil {
		options = &provider.CompleteOptions{}
	}

	prompt, err := p.template.Prompt(p.system, messages)

	if err != nil {
		return nil, err
	}

	req := &completionRequest{
		Stream: options.Stream != nil,

		Prompt: prompt,
		Stop:   []string{"[INST]"},

		Temperature: options.Temperature,
		TopP:        options.TopP,
		MinP:        options.MinP,
	}

	return req, nil
}

type embeddingRequest struct {
	Content string `json:"content"`
}

type embeddingResponse struct {
	Embedding []float32 `json:"embedding"`
}

type completionRequest struct {
	Prompt string `json:"prompt"`

	Stream bool     `json:"stream,omitempty"`
	Stop   []string `json:"stop,omitempty"`

	Temperature *float32 `json:"temperature,omitempty"`
	TopP        *float32 `json:"top_p,omitempty"`
	MinP        *float32 `json:"min_p,omitempty"`
}

type completionResponse struct {
	Model string `json:"model"`

	Prompt  string `json:"prompt"`
	Content string `json:"content"`

	Stop      bool `json:"stop"`
	Truncated bool `json:"truncated"`
}

func toCompletionReason(res completionResponse) provider.CompletionReason {
	if res.Truncated {
		return provider.CompletionReasonLength
	}

	if res.Stop {
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
