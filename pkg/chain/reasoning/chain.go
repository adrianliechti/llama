package reasoning

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"

	"github.com/adrianliechti/llama/pkg/chain"
	"github.com/adrianliechti/llama/pkg/provider"
)

var _ chain.Provider = &Chain{}

var (
	//go:embed system.txt
	systemPrompt string
)

type Chain struct {
	completer provider.Completer

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

	input := []provider.Message{
		{
			Role:    provider.MessageRoleSystem,
			Content: systemPrompt,
		},
	}

	for _, m := range messages {
		input = append(input, m)
	}

	// https://github.com/bklieger-groq/g1

	input = append(input, provider.Message{
		Role:    provider.MessageRoleAssistant,
		Content: "Thank you! I will now think step by step following my instructions, starting at the beginning after decomposing the problem.",
	})

	inputOptions := &provider.CompleteOptions{
		Temperature: options.Temperature,

		Format: provider.CompletionFormatJSON,
	}

	for {
		completion, err := c.completer.Complete(ctx, input, inputOptions)

		if err != nil {
			return nil, err
		}

		var step Step

		if err := json.Unmarshal([]byte(completion.Message.Content), &step); err != nil {
			return nil, err
		}

		println("Step:   ", step.Title)
		println("Result: ", step.Content)
		println()

		if step.NextAction == ActionFinalAnswer {
			break
		}

		input = append(input, provider.Message{
			Role:    provider.MessageRoleAssistant,
			Content: completion.Message.Content,
		})

		if len(input) > 25 {
			return nil, errors.New("too many steps")
		}
	}

	input = input[1:]

	input = append(input, provider.Message{
		Role:    provider.MessageRoleUser,
		Content: "Please provide the final answer based on your reasoning above.",
	})

	return c.completer.Complete(ctx, input, options)
}
