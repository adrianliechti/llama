package toolbox

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

func (c *Chain) Complete(ctx context.Context, messages []provider.Message, options *provider.CompleteOptions) (*provider.Completion, error) {
	if options == nil {
		options = new(provider.CompleteOptions)
	}

	functions := make(map[string]provider.Function)

	for _, f := range c.tools {
		functions[f.Name()] = provider.Function{
			Name:        f.Name(),
			Description: f.Description(),

			Parameters: f.Parameters(),
		}
	}

	for _, f := range options.Functions {
		functions[f.Name] = f
	}

	inputMessages := slices.Clone(messages)

	inputOptions := &provider.CompleteOptions{
		Functions: to.Values(functions),
	}

	for {
		completion, err := c.completer.Complete(ctx, inputMessages, inputOptions)

		if err != nil {
			return nil, err
		}

		var loop bool

		if completion.Reason == provider.CompletionReasonFunction {
			inputMessages = append(inputMessages, provider.Message{
				Role: provider.MessageRoleAssistant,

				Content:       completion.Message.Content,
				FunctionCalls: completion.Message.FunctionCalls,
			})

			for _, f := range completion.Message.FunctionCalls {
				tool, found := c.tools[f.Name]

				if !found {
					continue
				}

				var params map[string]any

				if err := json.Unmarshal([]byte(f.Arguments), &params); err != nil {
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

				inputMessages = append(inputMessages, provider.Message{
					Role: provider.MessageRoleFunction,

					Function: f.ID,
					Content:  string(data),
				})

				loop = true
			}
		}

		if !loop {
			return completion, nil
		}
	}
}
