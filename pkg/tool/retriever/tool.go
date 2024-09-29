package retriever

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
		name:        "retriever",
		description: "Query the knowledge base to find relevant documents to answer questions",

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

func (*Tool) Parameters() any {
	return map[string]any{
		"type": "object",

		"properties": map[string]any{
			"query": map[string]any{
				"type":        "string",
				"description": "The natural language query input. The query input should be clear and standalone",
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
			Title:    r.Title,
			Content:  r.Content,
			Location: r.Location,
		}

		results = append(results, result)
	}

	return results, nil
}
