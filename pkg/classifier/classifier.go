package classifier

import (
	"context"
	"errors"
	"regexp"

	"github.com/adrianliechti/llama/pkg/provider"
)

type Provider struct {
	completer provider.Completer

	classes []Class
}

type Class struct {
	Name string

	Description string
}

type Option func(*Provider)

func New(options ...Option) (*Provider, error) {
	p := &Provider{}

	for _, option := range options {
		option(p)
	}

	if p.completer == nil {
		return nil, errors.New("missing completer provider")
	}

	return p, nil
}

func WithCompleter(completer provider.Completer) Option {
	return func(p *Provider) {
		p.completer = completer
	}
}

func WithClasses(classes ...Class) Option {
	return func(p *Provider) {
		p.classes = classes
	}
}

func (p *Provider) Classify(ctx context.Context, input string) (string, error) {
	data := promptData{
		Input: input,

		Classes: p.classes,
	}

	prompt := executePromptTemplate(data)

	temperature := float32(0.1)

	completion, err := p.completer.Complete(ctx, []provider.Message{
		{
			Role:    provider.MessageRoleUser,
			Content: prompt,
		},
	}, &provider.CompleteOptions{
		Stop:        promptStop,
		Temperature: &temperature,
	})

	if err != nil {
		return "", err
	}

	return extractClass(completion.Message.Content)
}

func extractClass(s string) (string, error) {
	re := regexp.MustCompile(`Class: ([a-zA-Z]*)`)
	matches := re.FindAllStringSubmatch(s, -1)

	if len(matches) > 0 {
		match := matches[len(matches)-1]

		if len(match) == 2 {
			class := match[1]
			return class, nil
		}
	}

	return "", errors.New("no class found")
}
