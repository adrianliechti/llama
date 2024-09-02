package draw

import (
	"net/http"

	"github.com/adrianliechti/llama/pkg/provider"
)

type Option func(*Tool)

func WithRenderer(renderer provider.Renderer) Option {
	return func(t *Tool) {
		t.renderer = renderer
	}
}

func WithClient(client *http.Client) Option {
	return func(t *Tool) {
		t.client = client
	}
}
