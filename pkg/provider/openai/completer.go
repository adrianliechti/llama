package openai

import (
	"context"
	"encoding/base64"
	"io"
	"net/http"
	"strings"

	"github.com/adrianliechti/llama/pkg/provider"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/shared"
)

var _ provider.Completer = (*Completer)(nil)

type Completer struct {
	*Config
	completions *openai.ChatCompletionService
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
		Config:      cfg,
		completions: openai.NewChatCompletionService(cfg.Options()...),
	}, nil
}

func (c *Completer) Complete(ctx context.Context, messages []provider.Message, options *provider.CompleteOptions) (*provider.Completion, error) {
	if options == nil {
		options = new(provider.CompleteOptions)
	}

	req, err := c.convertCompletionRequest(messages, options)

	if err != nil {
		return nil, err
	}

	if options.Stream == nil {
		completion, err := c.completions.New(ctx, *req)

		if err != nil {
			return nil, convertError(err)
		}

		choice := completion.Choices[0]

		return &provider.Completion{
			ID:     completion.ID,
			Reason: toCompletionResult(choice.FinishReason),

			Message: provider.Message{
				Role:    provider.MessageRoleAssistant,
				Content: choice.Message.Content,

				ToolCalls: toToolCalls(choice.Message.ToolCalls),
			},

			Usage: &provider.Usage{
				InputTokens:  int(completion.Usage.PromptTokens),
				OutputTokens: int(completion.Usage.CompletionTokens),
			},
		}, nil
	} else {
		stream := c.completions.NewStreaming(ctx, *req)

		completion := openai.ChatCompletionAccumulator{}

		for stream.Next() {
			chunk := stream.Current()
			completion.AddChunk(chunk)

			if len(chunk.Choices) > 0 {
				reason := openai.ChatCompletionChoicesFinishReason(chunk.Choices[0].FinishReason)

				completion := provider.Completion{
					ID:     completion.ID,
					Reason: toCompletionResult(reason),

					Message: provider.Message{
						Role:    provider.MessageRoleAssistant,
						Content: chunk.Choices[0].Delta.Content,
					},
				}

				if err := options.Stream(ctx, completion); err != nil {
					return nil, err
				}
			}
		}

		if err := stream.Err(); err != nil {
			return nil, convertError(err)
		}

		choice := completion.Choices[0]

		return &provider.Completion{
			ID:     completion.ID,
			Reason: toCompletionResult(choice.FinishReason),

			Message: provider.Message{
				Role:    provider.MessageRoleAssistant,
				Content: choice.Message.Content,

				ToolCalls: toToolCalls(choice.Message.ToolCalls),
			},

			Usage: &provider.Usage{
				InputTokens:  int(completion.Usage.PromptTokens),
				OutputTokens: int(completion.Usage.CompletionTokens),
			},
		}, nil
	}
}

func (c *Completer) convertCompletionRequest(messages []provider.Message, options *provider.CompleteOptions) (*openai.ChatCompletionNewParams, error) {
	if options == nil {
		options = new(provider.CompleteOptions)
	}

	req := &openai.ChatCompletionNewParams{
		Model: openai.F(c.model),

		Tools:    openai.F([]openai.ChatCompletionToolParam{}),
		Messages: openai.F([]openai.ChatCompletionMessageParamUnion{}),
	}

	if options.Stream != nil && !strings.Contains(c.url, "openai.azure.com") {
		req.StreamOptions = openai.F(openai.ChatCompletionStreamOptionsParam{
			IncludeUsage: openai.F(true),
		})
	}

	if options.Format == provider.CompletionFormatJSON {
		req.ResponseFormat = openai.F[openai.ChatCompletionNewParamsResponseFormatUnion](shared.ResponseFormatJSONObjectParam{
			Type: openai.F(openai.ResponseFormatJSONObjectTypeJSONObject),
		})
	}

	if options.Stop != nil {
		stops := openai.ChatCompletionNewParamsStopArray(options.Stop)
		req.Stop = openai.F[openai.ChatCompletionNewParamsStopUnion](stops)
	}

	if options.MaxTokens != nil {
		req.MaxTokens = openai.F(int64(*options.MaxTokens))
	}

	if options.Temperature != nil {
		req.Temperature = openai.F(float64(*options.Temperature))
	}

	for _, m := range messages {
		switch m.Role {
		case provider.MessageRoleSystem:
			message := openai.SystemMessage(m.Content)
			req.Messages.Value = append(req.Messages.Value, message)

		case provider.MessageRoleUser:
			parts := []openai.ChatCompletionContentPartUnionParam{}

			if m.Content != "" {
				parts = append(parts, openai.TextPart(m.Content))
			}

			for _, f := range m.Files {
				data, err := io.ReadAll(f.Content)

				if err != nil {
					return nil, err
				}

				mime := http.DetectContentType(data)
				content := base64.StdEncoding.EncodeToString(data)

				url := "data:" + mime + ";base64," + content

				parts = append(parts, openai.ImagePart(url))
			}

			message := openai.UserMessageParts(parts...)
			req.Messages.Value = append(req.Messages.Value, message)

		case provider.MessageRoleAssistant:
			message := openai.AssistantMessage(m.Content)

			var toolcalls []openai.ChatCompletionMessageToolCallParam

			for _, t := range m.ToolCalls {
				toolcall := openai.ChatCompletionMessageToolCallParam{
					ID:   openai.F(t.ID),
					Type: openai.F(openai.ChatCompletionMessageToolCallTypeFunction),

					Function: openai.F(openai.ChatCompletionMessageToolCallFunctionParam{
						Name:      openai.F(t.Name),
						Arguments: openai.F(t.Arguments),
					}),
				}

				toolcalls = append(toolcalls, toolcall)
			}

			if len(toolcalls) > 0 {
				message.ToolCalls = openai.F(toolcalls)
			}

			req.Messages.Value = append(req.Messages.Value, message)

		case provider.MessageRoleTool:
			message := openai.ToolMessage(m.Tool, m.Content)
			req.Messages.Value = append(req.Messages.Value, message)
		}
	}

	for _, t := range options.Tools {
		tool := openai.ChatCompletionToolParam{
			Type: openai.F(openai.ChatCompletionToolTypeFunction),

			Function: openai.F(shared.FunctionDefinitionParam{
				Name:        openai.F(t.Name),
				Description: openai.F(t.Description),
				Parameters:  openai.F(shared.FunctionParameters(t.Parameters)),
				//Strict:      openai.F(true),
			}),
		}

		req.Tools.Value = append(req.Tools.Value, tool)
	}

	return req, nil
}

func toToolCalls(calls []openai.ChatCompletionMessageToolCall) []provider.ToolCall {
	var result []provider.ToolCall

	for _, c := range calls {
		if c.Function.Name != "" || c.Function.Arguments != "" {
			call := provider.ToolCall{
				ID: c.ID,

				Name:      c.Function.Name,
				Arguments: c.Function.Arguments,
			}

			result = append(result, call)
		}
	}

	return result
}

func toCompletionResult(val openai.ChatCompletionChoicesFinishReason) provider.CompletionReason {
	switch val {
	case openai.ChatCompletionChoicesFinishReasonStop:
		return provider.CompletionReasonStop

	case openai.ChatCompletionChoicesFinishReasonLength:
		return provider.CompletionReasonLength

	case openai.ChatCompletionChoicesFinishReasonToolCalls:
		return provider.CompletionReasonTool

	case openai.ChatCompletionChoicesFinishReasonContentFilter:
		return provider.CompletionReasonFilter

	default:
		return ""
	}
}
