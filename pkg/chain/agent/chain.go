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

	tools    []tool.Provider
	messages []provider.Message

	effort      provider.ReasoningEffort
	temperature *float32
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

func WithTools(tool ...tool.Provider) Option {
	return func(c *Chain) {
		c.tools = tool
	}
}

func WithEffort(effort provider.ReasoningEffort) Option {
	return func(c *Chain) {
		c.effort = effort
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

	if options.Effort == "" {
		options.Effort = c.effort
	}

	if options.Temperature == nil {
		options.Temperature = c.temperature
	}

	if len(c.messages) > 0 {
		values, err := template.Messages(c.messages, nil)

		if err != nil {
			return nil, err
		}

		messages = slices.Concat(values, messages)
	}

	input := slices.Clone(messages)

	agentTools := make(map[string]tool.Provider)
	inputTools := make(map[string]provider.Tool)

	for _, p := range c.tools {
		tools, err := p.Tools(ctx)

		if err != nil {
			return nil, err
		}

		for _, tool := range tools {
			agentTools[tool.Name] = p
			inputTools[tool.Name] = tool
		}
	}

	for _, t := range options.Tools {
		inputTools[t.Name] = t
	}

	var result *provider.Completion

	inputOptions := &provider.CompleteOptions{
		Effort: options.Effort,

		Stop:  options.Stop,
		Tools: to.Values(inputTools),

		MaxTokens:   options.MaxTokens,
		Temperature: options.Temperature,

		Format: options.Format,
		Schema: options.Schema,
	}

	var lastToolCallID string
	var lastToolCallName string

	streamToolCalls := map[string]provider.ToolCall{}

	stream := func(ctx context.Context, completion provider.Completion) error {
		for _, t := range completion.Message.ToolCalls {
			if t.ID != "" {
				lastToolCallID = t.ID
			}

			if t.Name != "" {
				lastToolCallName = t.Name
			}

			if lastToolCallName == "" {
				continue
			}

			if _, found := agentTools[lastToolCallName]; !found {
				call := streamToolCalls[lastToolCallID]
				call.ID = lastToolCallID
				call.Name = lastToolCallName
				call.Arguments += t.Arguments

				streamToolCalls[lastToolCallID] = call
			}
		}

		if completion.Message.Content != "" || completion.Reason != "" {
			completion.Message.ToolCalls = to.Values(streamToolCalls)

			return options.Stream(ctx, completion)
		}

		return nil
	}

	if options.Stream != nil {
		inputOptions.Stream = stream
	}

	for {
		completion, err := c.completer.Complete(ctx, input, inputOptions)

		if err != nil {
			return nil, err
		}

		input = append(input, completion.Message)

		var loop bool

		for _, t := range completion.Message.ToolCalls {
			p, found := agentTools[t.Name]

			if !found {
				continue
			}

			var params map[string]any

			if err := json.Unmarshal([]byte(t.Arguments), &params); err != nil {
				return nil, err
			}

			result, err := p.Execute(ctx, t.Name, params)

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
