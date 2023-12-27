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
	"time"

	"github.com/google/uuid"
	"github.com/sashabaranov/go-openai"
)

type Provider struct {
	client *http.Client

	url string

	model string

	system   string
	template PromptTemplate

	username string
	password string
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

func (p *Provider) Models(ctx context.Context) ([]openai.Model, error) {
	return []openai.Model{
		{
			ID: p.model,

			Object: "model",
			Root:   p.model,

			OwnedBy:   "owner",
			CreatedAt: time.Now().Unix(),
		},
	}, nil
}

func (p *Provider) Embed(ctx context.Context, request openai.EmbeddingRequest) (*openai.EmbeddingResponse, error) {
	input, err := convertEmbeddingRequest(request)

	if err != nil {
		return nil, err
	}

	list := &openai.EmbeddingResponse{
		Object: "list",
		Model:  request.Model,
	}

	for i, content := range input {
		req := &embeddingRequest{
			Content: strings.TrimSpace(content),
		}

		data, _ := json.Marshal(req)
		url, _ := url.JoinPath(p.url, "/embedding")

		r, _ := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(data))
		r.Header.Set("Content-Type", "application/json")
		r.Header.Set("Cache-Control", "no-cache")

		if p.password != "" {
			r.SetBasicAuth(p.username, p.password)
		}

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

		list.Data = append(list.Data, openai.Embedding{
			Index:  i,
			Object: "embedding",

			Embedding: result.Embedding,
		})
	}

	return list, nil
}

func (p *Provider) Complete(ctx context.Context, request openai.ChatCompletionRequest) (*openai.ChatCompletionResponse, error) {
	sessionID := uuid.New().String()

	req, err := p.convertCompletionRequest(request)

	if err != nil {
		return nil, err
	}

	data, _ := json.Marshal(req)
	url, _ := url.JoinPath(p.url, "/completion")

	r, _ := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(data))
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("Cache-Control", "no-cache")

	if p.password != "" {
		r.SetBasicAuth(p.username, p.password)
	}

	resp, err := p.client.Do(r)

	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("unable to complete chat")
	}

	defer resp.Body.Close()

	var result completionResponse

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	model := result.Model

	if model == "" {
		model = request.Model
	}

	if model == "" {
		model = p.model
	}

	content := p.template.RenderContent(result.Content)

	return &openai.ChatCompletionResponse{
		ID: sessionID,

		Model:  model,
		Object: "chat.completion",

		Created: time.Now().Unix(),

		Choices: []openai.ChatCompletionChoice{
			{
				Message: openai.ChatCompletionMessage{
					Content: content,
				},

				FinishReason: openai.FinishReasonStop,
			},
		},
	}, nil
}

func (p *Provider) CompleteStream(ctx context.Context, request openai.ChatCompletionRequest, stream chan<- openai.ChatCompletionStreamResponse) error {
	sessionID := uuid.New().String()

	req, err := p.convertCompletionRequest(request)

	if err != nil {
		return err
	}

	data, _ := json.Marshal(req)
	url, _ := url.JoinPath(p.url, "/completion")

	r, _ := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(data))
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("Accept", "text/event-stream")
	r.Header.Set("Connection", "keep-alive")
	r.Header.Set("Cache-Control", "no-cache")

	if err != nil {
		return err
	}

	resp, err := p.client.Do(r)

	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return errors.New("unable to complete chat")
	}

	reader := bufio.NewReader(resp.Body)

	for {
		data, err := reader.ReadBytes('\n')

		if err != nil {
			return err
		}

		data = bytes.TrimSpace(data)

		// if bytes.HasPrefix(data, errorPrefix) {
		// }

		data = bytes.TrimPrefix(data, headerData)

		if len(data) == 0 {
			continue
		}

		var result completionResponse

		if err := json.Unmarshal([]byte(data), &result); err != nil {
			return err
		}

		model := result.Model

		if model == "" {
			model = request.Model
		}

		if model == "" {
			model = p.model
		}

		status := openai.FinishReasonNull

		if result.Stop {
			status = openai.FinishReasonStop
		}

		content := p.template.RenderContent(result.Content)

		stream <- openai.ChatCompletionStreamResponse{
			ID: sessionID,

			Model:  model,
			Object: "chat.completion.chunk",

			Created: time.Now().Unix(),

			Choices: []openai.ChatCompletionStreamChoice{
				{
					Delta: openai.ChatCompletionStreamChoiceDelta{
						Role:    openai.ChatMessageRoleAssistant,
						Content: content,
					},

					FinishReason: status,
				},
			},
		}

		if result.Stop {
			break
		}
	}

	return nil
}

func convertEmbeddingRequest(request openai.EmbeddingRequest) ([]string, error) {
	data, _ := json.Marshal(request)

	type stringType struct {
		Input string `json:"input"`
	}

	var stringVal stringType

	if json.Unmarshal(data, &stringVal) == nil {
		if stringVal.Input != "" {
			return []string{stringVal.Input}, nil
		}
	}

	type sliceType struct {
		Input []string `json:"input"`
	}

	var sliceVal sliceType

	if json.Unmarshal(data, &sliceVal) == nil {
		if len(sliceVal.Input) > 0 {
			return sliceVal.Input, nil
		}
	}

	return nil, errors.New("invalid input format")
}

func (p *Provider) convertCompletionRequest(request openai.ChatCompletionRequest) (*completionRequest, error) {
	prompt, err := p.template.ConvertPrompt(p.system, request.Messages)

	if err != nil {
		return nil, err
	}

	result := &completionRequest{
		Stream: request.Stream,

		Prompt: prompt,
		Stop:   []string{"[INST]"},

		Temperature: request.Temperature,
		TopP:        request.TopP,

		NPredict: -1,
		//NPredict: 400,
	}

	if result.TopP == 0 {
		result.TopP = 0.9
		//result.TopP = 0.95
	}

	if result.Temperature == 0 {
		result.Temperature = 0.6
		//result.Temperature = 0.2
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
