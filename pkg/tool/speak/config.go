package speak

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

func WithSynthesizer(synthesizer provider.Synthesizer) Option {
	return func(t *Tool) {
		t.synthesizer = synthesizer
	}
}
