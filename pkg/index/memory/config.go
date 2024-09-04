package memory

import (
	"github.com/adrianliechti/llama/pkg/index"
)

type Config struct {
	embedder index.Embedder
}

type Option func(*Config)

func WithEmbedder(embedder index.Embedder) Option {
	return func(c *Config) {
		c.embedder = embedder
	}
}
