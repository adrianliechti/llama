package summarize

import (
	"context"
	"errors"
	"strings"

	"github.com/adrianliechti/llama/pkg/prompt"
	"github.com/adrianliechti/llama/pkg/provider"
)

var _ provider.Completer = &Provider{}

type Provider struct {
	prompt *prompt.Prompt

	completer provider.Completer
}

type Option func(*Provider)

func New(options ...Option) (*Provider, error) {
	p := &Provider{
		prompt: prompt.MustNew(promptTemplate),
	}

	for _, option := range options {
		option(p)
	}

	if p.completer == nil {
		return nil, errors.New("missing completer provider")
	}

	return p, nil
}

func WithPrompt(prompt *prompt.Prompt) Option {
	return func(p *Provider) {
		p.prompt = prompt
	}
}

func WithCompleter(completer provider.Completer) Option {
	return func(p *Provider) {
		p.completer = completer
	}
}

func (p *Provider) Complete(ctx context.Context, messages []provider.Message, options *provider.CompleteOptions) (*provider.Completion, error) {
	message := messages[len(messages)-1]

	if message.Role != provider.MessageRoleUser {
		return nil, errors.New("last message must be from user")
	}

	data := promptData{
		Input: strings.TrimSpace(message.Content),
	}

	prompt, err := p.prompt.Execute(data)

	if err != nil {
		return nil, err
	}

	println(prompt)

	message = provider.Message{
		Role:    provider.MessageRoleUser,
		Content: prompt,
	}

	messages = []provider.Message{message}

	return p.completer.Complete(ctx, messages, options)
}
