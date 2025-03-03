package tool

import (
	"context"
	"errors"

	"github.com/adrianliechti/wingman/pkg/provider"
)

type Tool = provider.Tool

var (
	ErrInvalidTool = errors.New("invalid tool")
)

type Provider interface {
	Tools(ctx context.Context) ([]Tool, error)
	Execute(ctx context.Context, name string, parameters map[string]any) (any, error)
}
