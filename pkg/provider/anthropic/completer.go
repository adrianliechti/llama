package anthropic

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"

	"github.com/adrianliechti/llama/pkg/provider"

	"github.com/anthropics/anthropic-sdk-go"
)

var _ provider.Completer = (*Completer)(nil)

type Completer struct {
	*Config
	messages *anthropic.MessageService
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
		Config:   cfg,
		messages: anthropic.NewMessageService(cfg.Options()...),
	}, nil
}

func (c *Completer) Complete(ctx context.Context, messages []provider.Message, options *provider.CompleteOptions) (*provider.Completion, error) {
	if options == nil {
		options = new(provider.CompleteOptions)
	}

	req, err := c.convertMessageRequest(messages, options)

	if err != nil {
		return nil, err
	}

	if options.Stream == nil {
		message, err := c.messages.New(ctx, *req)

		if err != nil {
			return nil, convertError(err)
		}

		return &provider.Completion{
			ID:     message.ID,
			Reason: toCompletionResult(message.StopReason),

			Message: provider.Message{
				Role:    provider.MessageRoleAssistant,
				Content: toContent(message.Content),

				ToolCalls: toToolCalls(message.Content),
			},

			Usage: &provider.Usage{
				InputTokens:  int(message.Usage.InputTokens),
				OutputTokens: int(message.Usage.OutputTokens),
			},
		}, nil
	} else {
		stream := c.messages.NewStreaming(ctx, *req)

		message := anthropic.Message{}

		for stream.Next() {
			event := stream.Current()
			message.Accumulate(event)

			switch delta := event.Delta.(type) {
			case anthropic.ContentBlockDeltaEventDelta:
				if delta.Text != "" {
					completion := provider.Completion{
						ID: message.ID,

						Message: provider.Message{
							Role:    provider.MessageRoleAssistant,
							Content: delta.Text,
						},
					}

					if err := options.Stream(ctx, completion); err != nil {
						return nil, err
					}
				}
			}
		}

		if err := stream.Err(); err != nil {
			return nil, convertError(err)
		}

		return &provider.Completion{
			ID:     message.ID,
			Reason: toCompletionResult(message.StopReason),

			Message: provider.Message{
				Role:    provider.MessageRoleAssistant,
				Content: toContent(message.Content),

				ToolCalls: toToolCalls(message.Content),
			},

			Usage: &provider.Usage{
				InputTokens:  int(message.Usage.InputTokens),
				OutputTokens: int(message.Usage.OutputTokens),
			},
		}, nil
	}
}

func (c *Completer) convertMessageRequest(input []provider.Message, options *provider.CompleteOptions) (*anthropic.MessageNewParams, error) {
	if options == nil {
		options = new(provider.CompleteOptions)
	}

	req := &anthropic.MessageNewParams{
		Model:     anthropic.F(c.model),
		MaxTokens: anthropic.F(int64(4096)),
	}

	var system []anthropic.TextBlockParam

	var tools []anthropic.ToolParam
	var messages []anthropic.MessageParam

	if options.Stop != nil {
		req.StopSequences = anthropic.F(options.Stop)
	}

	if options.MaxTokens != nil {
		req.MaxTokens = anthropic.F(int64(*options.MaxTokens))
	}

	if options.Temperature != nil {
		req.Temperature = anthropic.F(float64(*options.Temperature))
	}

	for _, m := range input {
		switch m.Role {
		case provider.MessageRoleSystem:
			system = append(system, anthropic.NewTextBlock(m.Content))

		case provider.MessageRoleUser:
			blocks := []anthropic.MessageParamContentUnion{}

			if m.Content != "" {
				blocks = append(blocks, anthropic.NewTextBlock(m.Content))
			}

			for _, f := range m.Files {
				data, err := io.ReadAll(f.Content)

				if err != nil {
					return nil, err
				}

				mime := http.DetectContentType(data)
				content := base64.StdEncoding.EncodeToString(data)

				blocks = append(blocks, anthropic.NewImageBlockBase64(mime, content))
			}

			message := anthropic.NewUserMessage(blocks...)
			messages = append(messages, message)

		case provider.MessageRoleAssistant:
			blocks := []anthropic.MessageParamContentUnion{}

			if m.Content != "" {
				blocks = append(blocks, anthropic.NewTextBlock(m.Content))
			}

			for _, t := range m.ToolCalls {
				var input any

				if err := json.Unmarshal([]byte(t.Arguments), &input); err != nil {
					input = t.Arguments
				}

				blocks = append(blocks, anthropic.NewToolUseBlockParam(t.ID, t.Name, input))
			}

			message := anthropic.NewAssistantMessage(blocks...)
			messages = append(messages, message)

		case provider.MessageRoleTool:
			message := anthropic.NewUserMessage(anthropic.NewToolResultBlock(m.Tool, m.Content, false))
			messages = append(messages, message)
		}
	}

	for _, t := range options.Tools {
		tool := anthropic.ToolParam{
			Name:        anthropic.F(t.Name),
			Description: anthropic.F(t.Description),
			InputSchema: anthropic.F[interface{}](t.Parameters),
		}

		tools = append(tools, tool)
	}

	if len(system) > 0 {
		req.System = anthropic.F(system)
	}

	if len(tools) > 0 {
		req.Tools = anthropic.F(tools)
	}

	if len(messages) > 0 {
		req.Messages = anthropic.F(messages)
	}

	return req, nil
}

func toContent(blocks []anthropic.ContentBlock) string {
	for _, b := range blocks {
		if b.Type != anthropic.ContentBlockTypeText {
			continue
		}

		return b.Text
	}

	return ""
}

func toToolCalls(blocks []anthropic.ContentBlock) []provider.ToolCall {
	var result []provider.ToolCall

	for _, b := range blocks {
		if b.Type != anthropic.ContentBlockTypeToolUse {
			continue
		}

		input, _ := b.Input.MarshalJSON()

		call := provider.ToolCall{
			ID: b.ID,

			Name:      b.Name,
			Arguments: string(input),
		}

		result = append(result, call)
	}

	return result
}

func toCompletionResult(val anthropic.MessageStopReason) provider.CompletionReason {
	switch val {
	case anthropic.MessageStopReasonEndTurn:
		return provider.CompletionReasonStop

	case anthropic.MessageStopReasonMaxTokens:
		return provider.CompletionReasonLength

	case anthropic.MessageStopReasonStopSequence:
		return provider.CompletionReasonStop

	case anthropic.MessageStopReasonToolUse:
		return provider.CompletionReasonTool

	default:
		return ""
	}
}
