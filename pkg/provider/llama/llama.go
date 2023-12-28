package llama

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strings"

	"github.com/adrianliechti/llama/pkg/provider"

	"github.com/sashabaranov/go-openai"
)

var (
	_ provider.Provider = &Provider{}
)

type Provider struct {
	client *http.Client

	url string

	model string

	system   string
	template PromptTemplate
}

type Option func(*Provider)

var (
	headerData = []byte("data: ")
	//errorPrefix = []byte(`data: {"error":`)
)

func New(options ...Option) *Provider {
	p := &Provider{
		client: http.DefaultClient,

		model:  "default",
		system: "",

		template: &PromptTemplateLLAMA{},
	}

	for _, option := range options {
		option(p)
	}

	return p
}

func WithClient(client *http.Client) Option {
	return func(p *Provider) {
		p.client = client
	}
}

func WithURL(url string) Option {
	return func(p *Provider) {
		p.url = url
	}
}

func WithModel(model string) Option {
	return func(p *Provider) {
		p.model = model
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

func (p *Provider) Models(ctx context.Context) ([]provider.Model, error) {
	return []provider.Model{
		{
			ID: p.model,
		},
	}, nil
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

func (p *Provider) Complete1(ctx context.Context, model string, messages []provider.Message, options *provider.CompleteOptions) (*provider.Message, error) {
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

		result := provider.Message{
			Role:    openai.ChatMessageRoleAssistant,
			Content: p.template.RenderContent(completion.Content),
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

		for {
			data, err := reader.ReadBytes('\n')

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

			options.Stream <- provider.Message{
				Role:    provider.MessageRoleAssistant,
				Content: p.template.RenderContent(completion.Content),
			}

			if completion.Stop {
				break
			}
		}

		result := provider.Message{
			Role:    provider.MessageRoleAssistant,
			Content: resultText.String(),
		}

		return &result, nil
	}
}

func (p *Provider) convertCompletionRequest(messages []provider.Message, options *provider.CompleteOptions) (*completionRequest, error) {
	if options == nil {
		options = &provider.CompleteOptions{}
	}

	prompt, err := p.template.ConvertPrompt(p.system, messages)

	if err != nil {
		return nil, err
	}

	result := &completionRequest{
		Stream: options.Stream != nil,

		Prompt: prompt,
		Stop:   []string{"[INST]"},

		//Temperature: request.Temperature,
		//TopP:        request.TopP,

		//NPredict: -1,
		////NPredict: 400,
	}

	// if result.TopP == 0 {
	// 	result.TopP = 0.9
	// 	//result.TopP = 0.95
	// }

	// if result.Temperature == 0 {
	// 	result.Temperature = 0.6
	// 	//result.Temperature = 0.2
	// }

	return result, nil
}

type embeddingRequest struct {
	Content string `json:"content"`
}

type embeddingResponse struct {
	Embedding []float32 `json:"embedding"`
}

type completionRequest struct {
	Stream bool `json:"stream"`

	Stop   []string `json:"stop"`
	Prompt string   `json:"prompt"`

	Temperature float32 `json:"temperature"`
	NPredict    int     `json:"n_predict"`
	TopP        float32 `json:"top_p"`
}

type completionResponse struct {
	Model string `json:"model"`

	Stop    bool   `json:"stop"`
	Prompt  string `json:"prompt"`
	Content string `json:"content"`

	Truncated bool `json:"truncated"`
}
