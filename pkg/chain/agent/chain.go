package agent

import (
	"context"
	"encoding/json"
	"errors"
	"slices"

	"github.com/adrianliechti/llama/pkg/chain"
	"github.com/adrianliechti/llama/pkg/provider"
	"github.com/adrianliechti/llama/pkg/to"
	"github.com/adrianliechti/llama/pkg/tool"
)

var _ chain.Provider = &Chain{}

type Chain struct {
	completer provider.Completer

	tools map[string]tool.Tool

	temperature *float32
}

type Option func(*Chain)

func New(options ...Option) (*Chain, error) {
	c := &Chain{
		tools: make(map[string]tool.Tool),
	}

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

func WithTools(tool ...tool.Tool) Option {
	return func(c *Chain) {
		for _, t := range tool {
			c.tools[t.Name()] = t
		}
	}
}

func WithTemperature(temperature float32) Option {
	return func(c *Chain) {
		c.temperature = &temperature
	}
}

func (c *Chain) Complete(ctx context.Context, messages []provider.Message, options *provider.CompleteOptions) (*provider.Completion, error) {
	if options == nil {
		options = new(provider.CompleteOptions)
	}

	if options.Temperature == nil {
		options.Temperature = c.temperature
	}

	tools := make(map[string]provider.Tool)

	for _, t := range c.tools {
		tools[t.Name()] = provider.Tool{
			Name:        t.Name(),
			Description: t.Description(),

			Parameters: t.Parameters(),
		}
	}

	for _, t := range options.Tools {
		tools[t.Name] = t
	}

	input := slices.Clone(messages)

	inputOptions := &provider.CompleteOptions{
		Temperature: options.Temperature,
		Tools:       to.Values(tools),
	}

	for {
		completion, err := c.completer.Complete(ctx, input, inputOptions)

		if err != nil {
			return nil, err
		}

		var loop bool

		if completion.Reason == provider.CompletionReasonFunction {
			input = append(input, provider.Message{
				Role: provider.MessageRoleAssistant,

				Content:   completion.Message.Content,
				ToolCalls: completion.Message.ToolCalls,
			})

			for _, t := range completion.Message.ToolCalls {
				tool, found := c.tools[t.Name]

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
		}

		if !loop {
			return completion, nil
		}
	}
}
