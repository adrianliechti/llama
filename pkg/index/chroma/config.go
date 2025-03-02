package chroma

import (
	"net/http"

	"github.com/adrianliechti/wingman/pkg/index"
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

func WithReranker(reranker index.Reranker) Option {
	return func(c *Client) {
		c.reranker = reranker
	}
}
