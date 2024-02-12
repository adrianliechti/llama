package llm

import (
	"context"
	"errors"
	"regexp"
	"strings"

	"github.com/adrianliechti/llama/pkg/classifier"
	"github.com/adrianliechti/llama/pkg/provider"
)

var _ classifier.Provider = &Provider{}

type Provider struct {
	completer provider.Completer

	classes []classifier.Class
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

	if len(p.classes) == 0 {
		return nil, errors.New("no classes provided")
	}

	return p, nil
}

func WithCompleter(completer provider.Completer) Option {
	return func(p *Provider) {
		p.completer = completer
	}
}

func WithClasses(classes ...classifier.Class) Option {
	return func(p *Provider) {
		p.classes = classes
	}
}

func (p *Provider) Classify(ctx context.Context, input string) (string, error) {
	data := promptData{
		Input: input,

		Classes: p.classes,
	}

	prompt, err := promptTemplate.Execute(data)

	if err != nil {
		return "", err
	}

	println(prompt)

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

	re := regexp.MustCompile(`[^a-z0-9:-_]+`)

	result := strings.ToLower(completion.Message.Content)
	result = re.ReplaceAllString(result, "")
	result = strings.ReplaceAll(result, "class:", "")
	result = strings.ReplaceAll(result, "category:", "")
	result = strings.TrimSpace(result)

	return result, nil
}
