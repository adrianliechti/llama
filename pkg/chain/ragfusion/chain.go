package ragfusion

import (
	"context"
	"errors"
	"strings"

	"github.com/adrianliechti/llama/pkg/chain"
	"github.com/adrianliechti/llama/pkg/classifier"
	"github.com/adrianliechti/llama/pkg/index"
	"github.com/adrianliechti/llama/pkg/prompt"
	"github.com/adrianliechti/llama/pkg/provider"
	"github.com/adrianliechti/llama/pkg/text"
	"github.com/adrianliechti/llama/pkg/to"
)

var _ chain.Provider = &Chain{}

type Chain struct {
	completer provider.Completer

	template *prompt.Template
	messages []provider.Message

	index index.Provider

	limit       *int
	temperature *float32

	filters map[string]classifier.Provider
}

type Option func(*Chain)

func New(options ...Option) (*Chain, error) {
	c := &Chain{
		template: prompt.MustTemplate(promptTemplate),

		filters: map[string]classifier.Provider{},
	}

	for _, option := range options {
		option(c)
	}

	if c.completer == nil {
		return nil, errors.New("missing completer provider")
	}

	if c.index == nil {
		return nil, errors.New("missing index provider")
	}

	return c, nil
}

func WithCompleter(completer provider.Completer) Option {
	return func(c *Chain) {
		c.completer = completer
	}
}

func WithTemplate(template *prompt.Template) Option {
	return func(c *Chain) {
		c.template = template
	}
}

func WithMessages(messages ...provider.Message) Option {
	return func(c *Chain) {
		c.messages = messages
	}
}

func WithIndex(index index.Provider) Option {
	return func(c *Chain) {
		c.index = index
	}
}

func WithLimit(limit int) Option {
	return func(c *Chain) {
		c.limit = &limit
	}
}

func WithTemperature(temperature float32) Option {
	return func(c *Chain) {
		c.temperature = &temperature
	}
}

func WithFilter(name string, classifier classifier.Provider) Option {
	return func(c *Chain) {
		c.filters[name] = classifier
	}
}

func (c *Chain) Complete(ctx context.Context, messages []provider.Message, options *provider.CompleteOptions) (*provider.Completion, error) {
	if options == nil {
		options = new(provider.CompleteOptions)
	}

	limit := 10

	if c.limit != nil {
		limit = *c.limit
	}

	if options.Temperature == nil {
		options.Temperature = c.temperature
	}

	message := messages[len(messages)-1]

	if message.Role != provider.MessageRoleUser {
		return nil, errors.New("last message must be from user")
	}

	filters := map[string]string{}

	for k, c := range c.filters {
		v, err := c.Classify(ctx, message.Content)

		if err != nil || v == "" {
			continue
		}

		filters[k] = v
	}

	queries, err := c.lala(ctx, message.Content)

	if err != nil {
		return nil, err
	}

	queries = append([]string{message.Content}, queries...)

	println(strings.Join(queries, "\n"))

	var list [][]index.Result

	for _, q := range queries {
		results, err := c.index.Query(ctx, q, &index.QueryOptions{
			Limit: to.Ptr(limit),

			Filters: filters,
		})

		if err != nil {
			return nil, err
		}

		list = append(list, results)
	}

	results := index.ReciprocalRankFusion(60, list...)
	results = results[:limit]

	data := promptData{
		Input: strings.TrimSpace(message.Content),
	}

	for _, r := range results {
		data.Results = append(data.Results, promptResult{
			Title:    r.Title,
			Content:  text.Normalize(r.Content),
			Location: r.Location,

			Metadata: r.Metadata,
		})
	}

	prompt, err := c.template.Execute(data)

	if err != nil {
		return nil, err
	}

	println(prompt)

	message = provider.Message{
		Role:    provider.MessageRoleUser,
		Content: prompt,
	}

	messages[len(messages)-1] = message

	return c.completer.Complete(ctx, messages, options)
}

func (c *Chain) lala(ctx context.Context, query string) ([]string, error) {
	messages := []provider.Message{
		{
			Role: provider.MessageRoleSystem,
			Content: "You are a helpful assistant that generates multiple search queries based on a single input query.\n" +
				"Answer with a plain list of 4 elements without no explaination and not formatting and no list style.",
		},
		{
			Role:    provider.MessageRoleUser,
			Content: "Generate multiple search queries related to: " + query,
		},
	}

	completion, err := c.completer.Complete(ctx, messages, &provider.CompleteOptions{})

	if err != nil {
		return nil, err
	}

	println(completion.Message.Content)

	queries := strings.Split(completion.Message.Content, "\n")
	return queries, nil
}
