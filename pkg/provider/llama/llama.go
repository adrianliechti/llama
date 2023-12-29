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

func New(options ...Option) (*Provider, error) {
	p := &Provider{
		client: http.DefaultClient,

		system: "",

		template: &PromptLLAMA{},
	}

	for _, option := range options {
		option(p)
	}

	if p.url == "" {
		return nil, errors.New("missing url")
	}

	return p, nil
}

func WithURL(url string) Option {
	return func(p *Provider) {
		p.url = url
	}
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
	req := &embeddingRequest{
		Content: strings.TrimSpace(content),
	}

	body, _ := json.Marshal(req)
	url, _ := url.JoinPath(p.url, "/embedding")

	r, _ := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("Cache-Control", "no-cache")

	resp, err := p.client.Do(r)

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

	req, err := p.convertCompletionRequest(messages, options)

	if err != nil {
		return nil, err
	}

	if options.Stream == nil {
		body, _ := json.Marshal(req)
		url, _ := url.JoinPath(p.url, "/completion")

		r, _ := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
		r.Header.Set("Content-Type", "application/json")
		r.Header.Set("Cache-Control", "no-cache")

		resp, err := p.client.Do(r)

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

		body, _ := json.Marshal(req)
		url, _ := url.JoinPath(p.url, "/completion")

		r, _ := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
		r.Header.Set("Content-Type", "application/json")
		r.Header.Set("Accept", "text/event-stream")
		r.Header.Set("Connection", "keep-alive")
		r.Header.Set("Cache-Control", "no-cache")

		resp, err := p.client.Do(r)

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

	result := &completionRequest{
		Stream: options.Stream != nil,

		Prompt: prompt,
		Stop:   []string{"[INST]"},
	}

	return result, nil
}

type embeddingRequest struct {
	Content string `json:"content"`
}

type embeddingResponse struct {
	Embedding []float32 `json:"embedding"`
}

type completionRequest struct {
	Prompt string `json:"prompt"`

	Stream bool     `json:"stream"`
	Stop   []string `json:"stop"`

	//Temperature float32 `json:"temperature"`
	//NPredict    int     `json:"n_predict"`
	//TopP        float32 `json:"top_p"`
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
