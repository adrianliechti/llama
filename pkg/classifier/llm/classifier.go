package llm

import (
	"context"
	"errors"
	"regexp"
	"strings"

	"github.com/adrianliechti/llama/pkg/classifier"
	"github.com/adrianliechti/llama/pkg/prompt"
	"github.com/adrianliechti/llama/pkg/provider"
)

var _ classifier.Provider = &Classifier{}

type Classifier struct {
	completer provider.Completer

	template *prompt.Template

	classes []classifier.Class
}

type Option func(*Classifier)

func New(options ...Option) (*Classifier, error) {
	p := &Classifier{
		template: prompt.MustTemplate(promptTemplate),
	}

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
	return func(c *Classifier) {
		c.completer = completer
	}
}

func WithTemplate(template *prompt.Template) Option {
	return func(c *Classifier) {
		c.template = template
	}
}

func WithClasses(classes ...classifier.Class) Option {
	return func(c *Classifier) {
		c.classes = classes
	}
}

func (c *Classifier) Classify(ctx context.Context, input string) (string, error) {
	data := promptData{
		Input: input,

		Classes: c.classes,
	}

	prompt, err := c.template.Execute(data)

	if err != nil {
		return "", err
	}

	println(prompt)

	temperature := float32(0.1)

	completion, err := c.completer.Complete(ctx, []provider.Message{
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
