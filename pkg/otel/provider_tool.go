package otel

import (
	"context"
	"strings"

	"github.com/adrianliechti/llama/pkg/tool"

	"go.opentelemetry.io/otel"
)

type Tool interface {
	Observable
	tool.Tool
}

type observableTool struct {
	name    string
	library string

	provider string

	tool tool.Tool
}

func NewTool(provider string, p tool.Tool) Tool {
	library := strings.ToLower(provider)

	return &observableTool{
		tool: p,

		name:    strings.TrimSuffix(strings.ToLower(provider), "-tool") + "-tool",
		library: library,

		provider: provider,
	}
}

func (p *observableTool) otelSetup() {
}

func (p *observableTool) Name() string {
	return p.tool.Name()
}

func (p *observableTool) Description() string {
	return p.tool.Description()
}

func (p *observableTool) Parameters() any {
	return p.tool.Parameters()
}

func (p *observableTool) Execute(ctx context.Context, parameters map[string]any) (any, error) {
	ctx, span := otel.Tracer(p.library).Start(ctx, p.name)
	defer span.End()

	result, err := p.tool.Execute(ctx, parameters)

	return result, err
}
