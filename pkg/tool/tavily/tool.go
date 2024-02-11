package tavily

import (
	"context"
	"errors"

	"github.com/adrianliechti/llama/pkg/index/tavily"
	"github.com/adrianliechti/llama/pkg/jsonschema"
	"github.com/adrianliechti/llama/pkg/tool"
)

var _ tool.Tool = &Tool{}

type Tool struct {
	client *tavily.Client
}

func New(token string) (*Tool, error) {
	client, err := tavily.New(token)

	if err != nil {
		return nil, err
	}

	t := &Tool{
		client: client,
	}

	return t, nil
}

func (*Tool) Name() string {
	return "tavily_search"
}

func (*Tool) Description() string {
	return "Get information on recent events from the web."
}

func (*Tool) Parameters() jsonschema.Definition {
	return jsonschema.Definition{
		Type: jsonschema.DataTypeObject,

		Properties: map[string]jsonschema.Definition{
			"query": {
				Type:        jsonschema.DataTypeString,
				Description: "The search query to use. For example: 'Latest news on Nvidia stock performance'",
			},
		},

		Required: []string{"query"},
	}
}

func (t *Tool) Execute(ctx context.Context, parameters map[string]any) (any, error) {
	val, ok := parameters["query"]

	if !ok {
		return nil, errors.New("missing query parameter")
	}

	query, ok := val.(string)

	if !ok {
		return nil, errors.New("invalid query parameter")
	}

	results, err := t.client.Query(ctx, query, nil)

	if err != nil {
		return nil, err
	}

	return results, nil
}
