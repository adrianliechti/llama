package provider

import (
	"context"
)

type Embedder interface {
	Embed(ctx context.Context, input string) (*Embedding, error)
}

type Embedding struct {
	Data []float32

	Usage *Usage
}
