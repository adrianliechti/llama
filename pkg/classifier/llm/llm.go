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

	categories []Category
}

type Category struct {
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

func WithCategories(categories ...Category) Option {
	return func(p *Provider) {
		p.categories = categories
	}
}

func (p *Provider) Categorize(ctx context.Context, input string) (string, error) {
	data := promptData{
		Input: input,

		Categories: p.categories,
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

	class := strings.TrimSpace(completion.Message.Content)
	class = strings.ToLower(class)
	class = strings.ReplaceAll(class, "class:", "")
	class = strings.ReplaceAll(class, "category:", "")

	return extractClass(class)
}

func extractClass(s string) (string, error) {
	re := regexp.MustCompile(`([a-zA-Z]*).*`)
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
