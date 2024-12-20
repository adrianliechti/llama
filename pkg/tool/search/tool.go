package search

import (
	"context"
	"errors"

	"github.com/adrianliechti/llama/pkg/index"
	"github.com/adrianliechti/llama/pkg/tool"
)

var _ tool.Tool = &Tool{}

type Tool struct {
	name        string
	description string

	index index.Provider
}

func New(index index.Provider, options ...Option) (*Tool, error) {
	t := &Tool{
		name:        "search",
		description: "Search online if the requested information cannot be found in the language model or the information could be present in a time after the language model was trained",

		index: index,
	}

	for _, option := range options {
		option(t)
	}

	if t.index == nil {
		return nil, errors.New("missing index provider")
	}

	return t, nil
}

func (t *Tool) Name() string {
	return t.name
}

func (t *Tool) Description() string {
	return t.description
}

func (*Tool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",

		"properties": map[string]any{
			"query": map[string]any{
				"type":        "string",
				"description": "the text to search online for to get the necessary information",
			},
		},

		"required": []string{"query"},
	}
}

func (t *Tool) Execute(ctx context.Context, parameters map[string]any) (any, error) {
	query, ok := parameters["query"].(string)

	if !ok {
		return nil, errors.New("missing query parameter")
	}

	options := &index.QueryOptions{}

	data, err := t.index.Query(ctx, query, options)

	if err != nil {
		return nil, err
	}

	results := []Result{}

	for _, r := range data {
		result := Result{
			Title:   r.Title,
			Content: r.Content,

			URL: r.Location,
		}

		results = append(results, result)
	}

	return results, nil
}
