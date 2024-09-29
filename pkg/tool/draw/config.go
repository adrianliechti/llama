package draw

import (
	"net/http"

	"github.com/adrianliechti/llama/pkg/provider"
)

type Option func(*Tool)

func WithClient(client *http.Client) Option {
	return func(t *Tool) {
		t.client = client
	}
}

func WithName(val string) Option {
	return func(t *Tool) {
		t.name = val
	}
}

func WithDescription(val string) Option {
	return func(t *Tool) {
		t.description = val
	}
}

func WithRenderer(renderer provider.Renderer) Option {
	return func(t *Tool) {
		t.renderer = renderer
	}
}
