package openai

import (
	"context"
	"errors"
	"io"
	"strings"

	"github.com/adrianliechti/llama/pkg/provider"
	"github.com/sashabaranov/go-openai"
)

var (
	_ provider.Provider = &Provider{}
)

var (
	ErrInvalidModelMapping = errors.New("invalid model mapping")
)

type Provider struct {
	url   string
	token string

	client *openai.Client
}

type Option func(*Provider)

func New(options ...Option) (*Provider, error) {
	p := &Provider{}

	for _, option := range options {
		option(p)
	}

	config := openai.DefaultConfig(p.token)

	if p.url != "" {
		config.BaseURL = p.url
	}

	if strings.Contains(p.url, "openai.azure.com") {
		config = openai.DefaultAzureConfig(p.token, p.url)
	}

	p.client = openai.NewClientWithConfig(config)

	return p, nil
}

func WithURL(url string) Option {
	return func(p *Provider) {
		p.url = url
	}
}

func WithToken(token string) Option {
	return func(p *Provider) {
		p.token = token
	}
}

func (p *Provider) Embed(ctx context.Context, model, content string) ([]float32, error) {
	// if p.mapper != nil {
	// 	model = p.mapper.To(model)
	// }

	req := openai.EmbeddingRequest{
		Input: content,
		Model: openai.AdaEmbeddingV2,
	}

	result, err := p.client.CreateEmbeddings(ctx, req)

	if err != nil {
		return nil, err
	}

	return result.Data[0].Embedding, nil
}

func (p *Provider) Complete(ctx context.Context, model string, messages []provider.Message, options *provider.CompleteOptions) (*provider.Completion, error) {
	if options == nil {
		options = &provider.CompleteOptions{}
	}

	if model == "" {
		return nil, ErrInvalidModelMapping
	}

	req, err := convertCompletionRequest(model, messages, options)

	if err != nil {
		return nil, err
	}

	if options.Stream == nil {
		completion, err := p.client.CreateChatCompletion(ctx, *req)

		if err != nil {
			return nil, err
		}

		choice := completion.Choices[0]

		result := provider.Completion{
			Message: &provider.Message{
				Role:    toMessageRole(choice.Message.Role),
				Content: choice.Message.Content,
			},

			Reason: toCompletionResult(choice.FinishReason),
		}

		return &result, nil

	} else {
		defer close(options.Stream)

		stream, err := p.client.CreateChatCompletionStream(ctx, *req)

		if err != nil {
			return nil, err
		}

		var resultText strings.Builder
		var resultRole provider.MessageRole
		var resultReason provider.CompletionReason

		for {
			completion, err := stream.Recv()

			if errors.Is(err, io.EOF) {
				break
			}

			if err != nil {
				return nil, err
			}

			choice := completion.Choices[0]

			resultText.WriteString(choice.Delta.Content)

			resultRole = toMessageRole(choice.Delta.Role)
			resultReason = toCompletionResult(choice.FinishReason)

			options.Stream <- provider.Completion{
				Message: &provider.Message{
					Role:    resultRole,
					Content: choice.Delta.Content,
				},

				Reason: resultReason,
			}

			if choice.FinishReason != "" {
				if choice.FinishReason == openai.FinishReasonStop {
					break
				}

				return nil, errors.New("unexpected finish reason: " + string(choice.FinishReason))
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

func convertCompletionRequest(model string, messages []provider.Message, options *provider.CompleteOptions) (*openai.ChatCompletionRequest, error) {
	if options == nil {
		options = &provider.CompleteOptions{}
	}

	req := &openai.ChatCompletionRequest{
		Model: model,
	}

	for _, m := range messages {
		req.Messages = append(req.Messages, openai.ChatCompletionMessage{
			Role:    convertMessageRole(m.Role),
			Content: m.Content,
		})
	}

	return req, nil
}

func convertMessageRole(r provider.MessageRole) string {
	switch r {

	case provider.MessageRoleSystem:
		return openai.ChatMessageRoleSystem

	case provider.MessageRoleUser:
		return openai.ChatMessageRoleUser

	case provider.MessageRoleAssistant:
		return openai.ChatMessageRoleAssistant

	default:
		return ""
	}
}

func toMessageRole(val string) provider.MessageRole {
	switch val {
	case openai.ChatMessageRoleSystem:
		return provider.MessageRoleSystem

	case openai.ChatMessageRoleUser:
		return provider.MessageRoleUser

	case openai.ChatMessageRoleAssistant:
		return provider.MessageRoleAssistant

	// case openai.ChatMessageRoleFunction:
	// 	return provider.MessageRoleFunction

	// case openai.ChatMessageRoleTool:
	// 	return provider.MessageRoleTool

	default:
		return provider.MessageRoleAssistant
	}
}

func toCompletionResult(val openai.FinishReason) provider.CompletionReason {
	switch val {
	case openai.FinishReasonStop:
		return provider.CompletionReasonStop

	case openai.FinishReasonLength:
		return provider.CompletionReasonLength

	default:
		return ""
	}
}
