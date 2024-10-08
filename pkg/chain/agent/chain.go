package agent

import (
	"context"
	"encoding/json"
	"errors"
	"slices"

	"github.com/adrianliechti/llama/pkg/chain"
	"github.com/adrianliechti/llama/pkg/provider"
	"github.com/adrianliechti/llama/pkg/template"
	"github.com/adrianliechti/llama/pkg/to"
	"github.com/adrianliechti/llama/pkg/tool"
)

var _ chain.Provider = &Chain{}

type Chain struct {
	completer provider.Completer

	tools    []tool.Tool
	messages []provider.Message
}

type Option func(*Chain)

func New(options ...Option) (*Chain, error) {
	c := &Chain{}

	for _, option := range options {
		option(c)
	}

	if c.completer == nil {
		return nil, errors.New("missing completer provider")
	}

	return c, nil
}

func WithCompleter(completer provider.Completer) Option {
	return func(c *Chain) {
		c.completer = completer
	}
}

func WithMessages(messages ...provider.Message) Option {
	return func(c *Chain) {
		c.messages = messages
	}
}

func WithTools(tool ...tool.Tool) Option {
	return func(c *Chain) {
		c.tools = tool
	}
}

func (c *Chain) Complete(ctx context.Context, messages []provider.Message, options *provider.CompleteOptions) (*provider.Completion, error) {
	if options == nil {
		options = new(provider.CompleteOptions)
	}

	if len(c.messages) > 0 {
		values, err := template.ApplyMessages(c.messages, nil)

		if err != nil {
			return nil, err
		}

		messages = slices.Concat(values, messages)
	}

	input := slices.Clone(messages)

	agentTools := make(map[string]tool.Tool)
	inputTools := make(map[string]provider.Tool)

	for _, t := range c.tools {
		tool := provider.Tool{
			Name:        t.Name(),
			Description: t.Description(),

			Parameters: t.Parameters(),
		}

		agentTools[tool.Name] = t
		inputTools[tool.Name] = tool
	}

	for _, t := range options.Tools {
		inputTools[t.Name] = t
	}

	var result *provider.Completion

	for {
		inputOptions := &provider.CompleteOptions{
			Stream: options.Stream,

			Tools: to.Values(inputTools),

			MaxTokens:   options.MaxTokens,
			Temperature: options.Temperature,

			Format: options.Format,
		}

		completion, err := c.completer.Complete(ctx, input, inputOptions)

		if err != nil {
			return nil, err
		}

		input = append(input, completion.Message)

		var loop bool

		for _, t := range completion.Message.ToolCalls {
			tool, found := agentTools[t.Name]

			if !found {
				continue
			}

			var params map[string]any

			if err := json.Unmarshal([]byte(t.Arguments), &params); err != nil {
				return nil, err
			}

			result, err := tool.Execute(ctx, params)

			if err != nil {
				return nil, err
			}

			data, err := json.Marshal(result)

			if err != nil {
				return nil, err
			}

			input = append(input, provider.Message{
				Role: provider.MessageRoleTool,

				Tool:    t.ID,
				Content: string(data),
			})

			loop = true
		}

		if !loop {
			result = completion
			break
		}
	}

	if result == nil {
		return nil, errors.New("unable to handle request")
	}

	return result, nil
}
