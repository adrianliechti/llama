package qdrant

import (
	"net/http"

	"github.com/adrianliechti/llama/pkg/index"
)

type Option func(*Client)

func WithClient(client *http.Client) Option {
	return func(c *Client) {
		c.client = client
	}
}

func WithEmbedder(embedder index.Embedder) Option {
	return func(c *Client) {
		c.embedder = embedder
	}
}
