package memory

import (
	"context"
	"errors"

	"github.com/adrianliechti/llama/pkg/tool"
)

var _ tool.Tool = &Tool{}

type Tool struct {
	name        string
	description string
}

func New(options ...Option) (*Tool, error) {
	t := &Tool{
		name:        "memory",
		description: "The `memory` tool allows you to persist information across conversations. Address your message `to=memory` and write whatever information you want to remember. The information will appear in the model set context below in future conversations.",
	}

	for _, option := range options {
		option(t)
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
				"claim":       "string",
				"description": "the information to persist across conversations",
			},
		},

		"required": []string{"claim"},
	}
}

func (t *Tool) Execute(ctx context.Context, parameters map[string]any) (any, error) {
	claim, ok := parameters["claim"].(string)

	if !ok {
		return nil, errors.New("missing claim parameter")
	}

	_ = claim

	println(claim)

	return "ok", nil
}
