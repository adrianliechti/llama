package memory

import (
	"github.com/adrianliechti/llama/pkg/index"
)

type Option func(*Provider)

func WithEmbedder(embedder index.Embedder) Option {
	return func(p *Provider) {
		p.embedder = embedder
	}
}
