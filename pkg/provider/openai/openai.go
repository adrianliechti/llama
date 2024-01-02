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

	req, err := p.convertCompletionRequest(model, messages, options)

	if err != nil {
		return nil, err
	}

	if options.Stream == nil {
		completion, err := p.client.CreateChatCompletion(ctx, *req)

		if err != nil {
			var oaierr *openai.APIError

			if errors.As(err, &oaierr) {
				return nil, errors.New(oaierr.Message)
			}

			return nil, err
		}

		choice := completion.Choices[0]

		return &provider.Completion{
			ID: completion.ID,

			Reason: toCompletionResult(choice.FinishReason),

			Message: provider.Message{
				Role:    toMessageRole(choice.Message.Role),
				Content: choice.Message.Content,
			},

			Functions: toFunctionCalls(choice.Message.ToolCalls),
		}, nil
	} else {
		defer close(options.Stream)

		stream, err := p.client.CreateChatCompletionStream(ctx, *req)

		if err != nil {
			return nil, err
		}

		var resultID string
		var resultText strings.Builder
		var resultRole provider.MessageRole
		var resultReason provider.CompletionReason
		var resultFunctions []provider.FunctionCall

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

			resultID = completion.ID
			resultRole = toMessageRole(choice.Delta.Role)
			resultReason = toCompletionResult(choice.FinishReason)
			resultFunctions = toFunctionCalls(choice.Delta.ToolCalls)

			options.Stream <- provider.Completion{
				ID: completion.ID,

				Reason: resultReason,

				Message: provider.Message{
					Role:    resultRole,
					Content: choice.Delta.Content,
				},

				Functions: resultFunctions,
			}

			if choice.FinishReason != "" {
				break
			}
		}

		result := provider.Completion{
			ID: resultID,

			Reason: resultReason,

			Message: provider.Message{
				Role:    resultRole,
				Content: resultText.String(),
			},

			Functions: resultFunctions,
		}

		return &result, nil
	}
}

func (p *Provider) convertCompletionRequest(model string, messages []provider.Message, options *provider.CompleteOptions) (*openai.ChatCompletionRequest, error) {
	if options == nil {
		options = &provider.CompleteOptions{}
	}

	req := &openai.ChatCompletionRequest{
		Model: model,
	}

	if options.Format == provider.CompletionFormatJSON {
		req.ResponseFormat = &openai.ChatCompletionResponseFormat{
			Type: openai.ChatCompletionResponseFormatTypeJSONObject,
		}
	}

	for _, f := range options.Functions {
		tool := openai.Tool{
			Type: openai.ToolTypeFunction,

			Function: openai.FunctionDefinition{
				Name:       f.Name,
				Parameters: f.Parameters,

				Description: f.Description,
			},
		}

		req.Tools = append(req.Tools, tool)
	}

	if options.Temperature != nil {
		req.Temperature = *options.Temperature
	}

	if options.TopP != nil {
		req.TopP = *options.TopP
	}

	for _, m := range messages {
		message := openai.ChatCompletionMessage{
			Role:    convertMessageRole(m.Role),
			Content: m.Content,

			ToolCallID: m.FunctionID,
		}

		for _, f := range m.FunctionCalls {
			call := openai.ToolCall{
				ID:   f.ID,
				Type: openai.ToolTypeFunction,

				Function: openai.FunctionCall{
					Name:      f.Name,
					Arguments: f.Arguments,
				},
			}

			message.ToolCalls = append(message.ToolCalls, call)
		}

		req.Messages = append(req.Messages, message)
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

	case provider.MessageRoleFunction:
		return openai.ChatMessageRoleTool

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

func toFunctionCalls(calls []openai.ToolCall) []provider.FunctionCall {
	var result []provider.FunctionCall

	for _, c := range calls {
		if c.Type == openai.ToolTypeFunction {
			result = append(result, provider.FunctionCall{
				ID: c.ID,

				Name:      c.Function.Name,
				Arguments: c.Function.Arguments,
			})
		}
	}

	return result
}

func toCompletionResult(val openai.FinishReason) provider.CompletionReason {
	switch val {
	case openai.FinishReasonStop:
		return provider.CompletionReasonStop

	case openai.FinishReasonLength:
		return provider.CompletionReasonLength

	case openai.FinishReasonToolCalls:
		return provider.CompletionReasonFunction

	default:
		return ""
	}
}
