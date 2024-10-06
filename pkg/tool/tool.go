package tool

import (
	"context"
)

type Tool interface {
	Name() string
	Description() string

	Parameters() map[string]any

	Execute(ctx context.Context, parameters map[string]any) (any, error)
}
