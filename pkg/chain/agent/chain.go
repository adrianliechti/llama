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

	messages []provider.Message

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

func WithMessages(messages ...provider.Message) Option {
	return func(c *Chain) {
		c.messages = messages
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

	if len(c.messages) > 0 {
		values, err := template.ApplyMessages(c.messages, nil)

		if err != nil {
			return nil, err
		}

		messages = slices.Concat(values, messages)
	}

	input := slices.Clone(messages)

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

	var result *provider.Completion

	for {
		inputOptions := &provider.CompleteOptions{
			Temperature: options.Temperature,
			Tools:       to.Values(tools),
		}

		calls := make(map[string]*provider.ToolCall)

		done := make(chan any)

		if options.Stream != nil {
			stream := make(chan provider.Completion)

			inputOptions.Stream = stream

			var toolID string
			var toolName string

			go func() {
				for m := range stream {
					if m.Reason != "" || m.Message.Content != "" {
						toolID = ""
						toolName = ""
					}

					var tool bool

					for _, t := range m.Message.ToolCalls {
						tool = true

						if t.ID != "" {
							toolID = t.ID
						}

						if t.Name != "" {
							toolName = t.Name
						}

						call, found := calls[toolID]

						if !found {
							call = &provider.ToolCall{
								ID:   toolID,
								Name: toolName,
							}

							calls[toolID] = call
						}

						call.Arguments = call.Arguments + t.Arguments
					}

					if !tool {
						options.Stream <- m
					}
				}

				done <- true
			}()
		}

		completion, err := c.completer.Complete(ctx, input, inputOptions)

		if err != nil {
			return nil, err
		}

		<-done

		for _, c := range completion.Message.ToolCalls {
			calls[c.ID] = &provider.ToolCall{
				ID: c.ID,

				Name:      c.Name,
				Arguments: c.Arguments,
			}
		}

		completion.Message.ToolCalls = nil

		for _, c := range calls {
			completion.Message.ToolCalls = append(completion.Message.ToolCalls, *c)
		}

		var loop bool

		if len(completion.Message.ToolCalls) > 0 {
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
			result = completion
			break
		}
	}

	if options.Stream != nil {
		close(options.Stream)
	}

	if result == nil {
		return nil, errors.New("unable to handle request")
	}

	return result, nil
}
