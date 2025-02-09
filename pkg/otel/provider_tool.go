package otel

import (
	"context"
	"strings"

	"github.com/adrianliechti/llama/pkg/tool"

	"go.opentelemetry.io/otel"
)

type Tool interface {
	Observable
	tool.Provider
}

type observableTool struct {
	name    string
	library string

	provider string

	tool tool.Provider
}

func NewTool(provider string, p tool.Provider) Tool {
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

func (p *observableTool) Tools(ctx context.Context) ([]tool.Tool, error) {
	ctx, span := otel.Tracer(p.library).Start(ctx, p.name)
	defer span.End()

	tools, err := p.tool.Tools(ctx)

	return tools, err
}

func (p *observableTool) Execute(ctx context.Context, tool string, parameters map[string]any) (any, error) {
	ctx, span := otel.Tracer(p.library).Start(ctx, p.name)
	defer span.End()

	result, err := p.tool.Execute(ctx, tool, parameters)

	return result, err
}
