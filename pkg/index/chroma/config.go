package chroma

import (
	"net/http"

	"github.com/adrianliechti/llama/pkg/index"
)

type Config struct {
	client *http.Client

	url string

	embedder  index.Embedder
	namespace string
}

type Option func(*Config)

func WithClient(client *http.Client) Option {
	return func(c *Config) {
		c.client = client
	}
}

func WithEmbedder(embedder index.Embedder) Option {
	return func(c *Config) {
		c.embedder = embedder
	}
}
