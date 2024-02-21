package openai

import (
	"context"
	"encoding/base64"
	"errors"
	"io"
	"mime"
	"path/filepath"
	"strings"

	"github.com/adrianliechti/llama/pkg/provider"

	"github.com/sashabaranov/go-openai"
)

var _ provider.Completer = (*Completer)(nil)

type Completer struct {
	*Config
	client *openai.Client
}

func NewCompleter(options ...Option) (*Completer, error) {
	c := &Config{
		model: openai.GPT3Dot5Turbo,
	}

	for _, option := range options {
		option(c)
	}

	return &Completer{
		Config: c,
		client: c.Client(),
	}, nil
}

func (c *Completer) Complete(ctx context.Context, messages []provider.Message, options *provider.CompleteOptions) (*provider.Completion, error) {
	if options == nil {
		options = new(provider.CompleteOptions)
	}

	req, err := convertCompletionRequest(c.model, messages, options)

	if err != nil {
		return nil, err
	}

	if options.Stream == nil {
		completion, err := c.client.CreateChatCompletion(ctx, *req)

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
				Role:    provider.MessageRoleAssistant,
				Content: choice.Message.Content,

				FunctionCalls: toFunctionCalls(choice.Message.ToolCalls),
			},
		}, nil
	} else {
		defer close(options.Stream)

		stream, err := c.client.CreateChatCompletionStream(ctx, *req)

		if err != nil {
			return nil, err
		}

		var resultID string
		var resultText strings.Builder
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
			resultReason = toCompletionResult(choice.FinishReason)
			resultFunctions = toFunctionCalls(choice.Delta.ToolCalls)

			options.Stream <- provider.Completion{
				ID: completion.ID,

				Reason: resultReason,

				Message: provider.Message{
					Content: choice.Delta.Content,

					FunctionCalls: toFunctionCalls(choice.Delta.ToolCalls),
				},
			}

			if choice.FinishReason != "" {
				break
			}
		}

		result := provider.Completion{
			ID: resultID,

			Reason: resultReason,

			Message: provider.Message{
				Role:    provider.MessageRoleAssistant,
				Content: resultText.String(),

				FunctionCalls: resultFunctions,
			},
		}

		return &result, nil
	}
}

func convertCompletionRequest(model string, messages []provider.Message, options *provider.CompleteOptions) (*openai.ChatCompletionRequest, error) {
	if options == nil {
		options = new(provider.CompleteOptions)
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

	if options.Stop != nil {
		req.Stop = options.Stop
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

			ToolCallID: m.Function,
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
				mime := mime.TypeByExtension(filepath.Ext(f.Name))
				data, err := io.ReadAll(f.Content)

				if err != nil {
					return nil, err
				}

				content := base64.StdEncoding.EncodeToString(data)

				message.MultiContent = append(message.MultiContent, openai.ChatMessagePart{
					Type: openai.ChatMessagePartTypeImageURL,
					ImageURL: &openai.ChatMessageImageURL{
						URL: "data:" + mime + ";base64," + content,
					},
				})
			}
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
