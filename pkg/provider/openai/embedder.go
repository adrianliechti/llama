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
	cfg := &Config{
		model: string(openai.SmallEmbedding3),
	}

	for _, option := range options {
		option(cfg)
	}

	return &Embedder{
		Config: cfg,
		client: cfg.newClient(),
	}, nil
}

func (c *Embedder) Embed(ctx context.Context, content string) (*provider.Embedding, error) {
	req := openai.EmbeddingRequest{
		Input: content,
		Model: openai.EmbeddingModel(c.model),
	}

	result, err := c.client.CreateEmbeddings(ctx, req)

	if err != nil {
		convertError(err)
	}

	return &provider.Embedding{
		Data: result.Data[0].Embedding,

		Usage: &provider.Usage{
			InputTokens:  result.Usage.PromptTokens,
			OutputTokens: result.Usage.CompletionTokens,
		},
	}, nil
}
