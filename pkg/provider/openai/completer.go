package openai

import (
	"context"
	"encoding/base64"
	"errors"
	"io"
	"net/http"
	"strings"
	"unicode"

	"github.com/adrianliechti/llama/pkg/provider"

	"github.com/sashabaranov/go-openai"
)

var _ provider.Completer = (*Completer)(nil)

type Completer struct {
	*Config
	client *openai.Client
}

func NewCompleter(url, model string, options ...Option) (*Completer, error) {
	cfg := &Config{
		url:   url,
		model: model,
	}

	for _, option := range options {
		option(cfg)
	}

	return &Completer{
		Config: cfg,
		client: cfg.newClient(),
	}, nil
}

func (c *Completer) Complete(ctx context.Context, messages []provider.Message, options *provider.CompleteOptions) (*provider.Completion, error) {
	if options == nil {
		options = new(provider.CompleteOptions)
	}

	if c.limiter != nil {
		c.limiter.Wait(ctx)
	}

	req, err := c.convertCompletionRequest(messages, options)

	if err != nil {
		return nil, err
	}

	if options.Stream == nil {
		completion, err := c.client.CreateChatCompletion(ctx, *req)

		if err != nil {
			return nil, convertError(err)
		}

		choice := completion.Choices[0]

		return &provider.Completion{
			ID:     completion.ID,
			Reason: toCompletionResult(choice.FinishReason),

			Message: provider.Message{
				Role:    toMessageRole(choice.Message.Role),
				Content: choice.Message.Content,

				ToolCalls: toToolCalls(choice.Message.ToolCalls),
			},

			Usage: &provider.Usage{
				InputTokens:  completion.Usage.PromptTokens,
				OutputTokens: completion.Usage.CompletionTokens,
			},
		}, nil
	} else {
		defer close(options.Stream)

		stream, err := c.client.CreateChatCompletionStream(ctx, *req)

		if err != nil {
			return nil, err
		}

		result := provider.Completion{
			Message: provider.Message{
				Role: provider.MessageRoleAssistant,
			},
		}

		for {
			completion, err := stream.Recv()

			if err != nil {
				if errors.Is(err, io.EOF) {
					break
				}

				return nil, err
			}

			result.ID = completion.ID

			if completion.Usage != nil {
				result.Usage = &provider.Usage{
					InputTokens:  completion.Usage.PromptTokens,
					OutputTokens: completion.Usage.CompletionTokens,
				}
			}

			if len(completion.Choices) == 0 {
				continue
			}

			choice := completion.Choices[0]

			role := toMessageRole(choice.Delta.Role)

			if role == "" {
				role = provider.MessageRoleAssistant
			}

			content := choice.Delta.Content

			if result.Message.Content == "" {
				content = strings.TrimLeftFunc(content, unicode.IsSpace)
			}

			result.Reason = toCompletionResult(choice.FinishReason)

			result.Message.Role = role
			result.Message.Content += content
			result.Message.ToolCalls = toToolCalls(choice.Delta.ToolCalls)

			options.Stream <- provider.Completion{
				ID:     result.ID,
				Reason: result.Reason,

				Message: provider.Message{
					Role:    role,
					Content: content,

					ToolCalls: toToolCalls(choice.Delta.ToolCalls),
				},
			}
		}

		return &result, nil
	}
}

func (c *Completer) convertCompletionRequest(messages []provider.Message, options *provider.CompleteOptions) (*openai.ChatCompletionRequest, error) {
	if options == nil {
		options = new(provider.CompleteOptions)
	}

	req := &openai.ChatCompletionRequest{
		Model: c.model,
	}

	if options.Stream != nil {
		req.StreamOptions = &openai.StreamOptions{
			IncludeUsage: true,
		}
	}

	if options.Format == provider.CompletionFormatJSON {
		req.ResponseFormat = &openai.ChatCompletionResponseFormat{
			Type: openai.ChatCompletionResponseFormatTypeJSONObject,
		}
	}

	if options.Stop != nil {
		req.Stop = options.Stop
	}

	if options.MaxTokens != nil {
		req.MaxTokens = *options.MaxTokens
	}

	if options.Temperature != nil {
		req.Temperature = *options.Temperature
	}

	for _, t := range options.Tools {
		tool := openai.Tool{
			Type: openai.ToolTypeFunction,

			Function: &openai.FunctionDefinition{
				Name:       t.Name,
				Parameters: t.Parameters,

				Description: t.Description,
			},
		}

		req.Tools = append(req.Tools, tool)
	}

	for _, m := range messages {
		message := openai.ChatCompletionMessage{
			Role:    convertMessageRole(m.Role),
			Content: m.Content,

			ToolCallID: m.Tool,
		}

		if len(m.Files) > 0 {
			message.Content = ""

			message.MultiContent = []openai.ChatMessagePart{
				{
					Type: openai.ChatMessagePartTypeText,
					Text: m.Content,
				},
			}

			for _, f := range m.Files {
				data, err := io.ReadAll(f.Content)

				if err != nil {
					return nil, err
				}

				mime := http.DetectContentType(data)
				content := base64.StdEncoding.EncodeToString(data)

				message.MultiContent = append(message.MultiContent, openai.ChatMessagePart{
					Type: openai.ChatMessagePartTypeImageURL,
					ImageURL: &openai.ChatMessageImageURL{
						URL: "data:" + mime + ";base64," + content,
					},
				})
			}
		}

		for _, t := range m.ToolCalls {
			call := openai.ToolCall{
				ID:   t.ID,
				Type: openai.ToolTypeFunction,

				Function: openai.FunctionCall{
					Name:      t.Name,
					Arguments: t.Arguments,
				},
			}

			message.ToolCalls = append(message.ToolCalls, call)
		}

		req.Messages = append(req.Messages, message)
	}

	if strings.Contains(c.url, "openai.azure.com") {
		req.StreamOptions = nil
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

	case provider.MessageRoleTool:
		return openai.ChatMessageRoleTool

	default:
		return ""
	}
}

func toMessageRole(role string) provider.MessageRole {
	switch role {
	case openai.ChatMessageRoleSystem:
		return provider.MessageRoleSystem

	case openai.ChatMessageRoleUser:
		return provider.MessageRoleUser

	case openai.ChatMessageRoleAssistant:
		return provider.MessageRoleAssistant

	case openai.ChatMessageRoleTool:
		return provider.MessageRoleTool

	default:
		return ""
	}
}

func toToolCalls(calls []openai.ToolCall) []provider.ToolCall {
	var result []provider.ToolCall

	for _, c := range calls {
		if c.Type == openai.ToolTypeFunction {
			result = append(result, provider.ToolCall{
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
		return provider.CompletionReasonTool

	default:
		return ""
	}
}
