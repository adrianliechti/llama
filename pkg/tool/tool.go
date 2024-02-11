package tool

import (
	"context"

	"github.com/adrianliechti/llama/pkg/jsonschema"
)

type Tool interface {
	Name() string
	Description() string

	Parameters() jsonschema.Definition

	Execute(ctx context.Context, parameters map[string]any) (any, error)
}
