package openai

import (
	"context"

	"github.com/adrianliechti/llama/pkg/provider"

	"github.com/sashabaranov/go-openai"
)

var _ provider.Embedder = (*Embedder)(nil)

type Embedder struct {
	*Config
	client *openai.Client
}

func NewEmbedder(options ...Option) (*Embedder, error) {
	c := &Config{
		model: string(openai.AdaEmbeddingV2),
	}

	for _, option := range options {
		option(c)
	}

	return &Embedder{
		Config: c,
		client: c.Client(),
	}, nil
}

func (c *Embedder) Embed(ctx context.Context, content string) ([]float32, error) {
	req := openai.EmbeddingRequest{
		Input: content,
		Model: openai.EmbeddingModel(c.model),
	}

	result, err := c.client.CreateEmbeddings(ctx, req)

	if err != nil {
		return nil, err
	}

	return result.Data[0].Embedding, nil
}
