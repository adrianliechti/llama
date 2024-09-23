package speak

import (
	"net/http"

	"github.com/adrianliechti/llama/pkg/provider"
)

type Option func(*Tool)

func WithSynthesizer(synthesizer provider.Synthesizer) Option {
	return func(t *Tool) {
		t.synthesizer = synthesizer
	}
}

func WithClient(client *http.Client) Option {
	return func(t *Tool) {
		t.client = client
	}
}
