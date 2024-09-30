package provider

import (
	"context"
	"io"
)

type Renderer interface {
	Render(ctx context.Context, input string, options *RenderOptions) (*Image, error)
}

type RenderOptions struct {
}

type Image struct {
	ID string

	Name    string
	Content io.ReadCloser
}
