package search

import (
	"context"
	"errors"

	"github.com/adrianliechti/llama/pkg/index"
	"github.com/adrianliechti/llama/pkg/jsonschema"
	"github.com/adrianliechti/llama/pkg/tool"
)

var _ tool.Tool = &Tool{}

type Tool struct {
	name        string
	description string

	index index.Provider
}

func New(index index.Provider) (*Tool, error) {
	t := &Tool{
		name:        "search_online",
		description: "Search online if the requested information cannot be found in the language model or the information could be present in a time after the language model was trained.",

		index: index,
	}

	return t, nil
}

type Option func(*Tool)

func WithName(name string) Option {
	return func(t *Tool) {
		t.name = name
	}
}

func WithDescription(description string) Option {
	return func(t *Tool) {
		t.description = description
	}
}

func (t *Tool) Name() string {
	return t.name
}

func (t *Tool) Description() string {
	return t.description
}

func (*Tool) Parameters() jsonschema.Definition {
	return jsonschema.Definition{
		Type: jsonschema.DataTypeObject,

		Properties: map[string]jsonschema.Definition{
			"query": {
				Type:        jsonschema.DataTypeString,
				Description: "the text to search online to get the necessary information",
			},
		},

		Required: []string{"query"},
	}
}

func (t *Tool) Execute(ctx context.Context, parameters map[string]any) (any, error) {
	query, ok := parameters["query"].(string)

	if !ok {
		return nil, errors.New("missing query parameter")
	}

	documents, err := t.index.Query(ctx, query, &index.QueryOptions{})

	if err != nil {
		return nil, err
	}

	result := make([]Result, 0)

	for _, d := range documents {
		result = append(result, Result{
			Title:    d.Title,
			Content:  d.Content,
			Location: d.Location,
		})
	}

	return result, nil
}

type Result struct {
	Title    string
	Content  string
	Location string
}
