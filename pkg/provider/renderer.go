package provider

import (
	"context"
)

type Renderer interface {
	Render(ctx context.Context, input string, options *RenderOptions) (*Image, error)
}

type RenderOptions struct {
}
