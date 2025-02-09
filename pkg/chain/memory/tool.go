package memory

import (
	"context"
	"errors"

	"github.com/adrianliechti/llama/pkg/tool"
)

type Tool struct {
	Claims []string
}

func (t *Tool) Tools(ctx context.Context) ([]tool.Tool, error) {
	return []tool.Tool{
		{
			Name:        "memory",
			Description: "The `memory` tool allows you to persist information across conversations. The information will appear in the model set context below in future conversations.",

			Parameters: map[string]any{
				"type": "object",

				"properties": map[string]any{
					"claim": map[string]any{
						"type":        "string",
						"description": "the information to persist across conversations",
					},
				},

				"required": []string{"claim"},
			},
		},
	}, nil
}

func (t *Tool) Execute(ctx context.Context, name string, parameters map[string]any) (any, error) {
	if name != "memory" {
		return nil, tool.ErrInvalidTool
	}

	claim, ok := parameters["claim"].(string)

	if !ok {
		return nil, errors.New("missing claim parameter")
	}

	t.Claims = append(t.Claims, claim)

	result := map[string]any{
		"status": "ok",
	}

	return result, nil
}
