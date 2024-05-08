package provider

import (
	"context"
)

type Embedder interface {
	Embed(ctx context.Context, content string) (Embeddings, error)
}

type Embeddings []float32
