package provider

import (
	"context"
	"io"
)

type Renderer interface {
	Render(ctx context.Context, input string, options *RenderOptions) (*Image, error)
}

type RenderOptions struct {
	Style ImageStyle
}

type ImageStyle string

const (
	ImageStyleNatural ImageStyle = "natural"
	ImageStyleVivid   ImageStyle = "vivid"
)

type Image struct {
	ID string

	Name    string
	Content io.ReadCloser
}
