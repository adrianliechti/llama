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

func NewEmbedder(url, model string, options ...Option) (*Embedder, error) {
	cfg := &Config{
		url:   url,
		model: model,
	}

	for _, option := range options {
		option(cfg)
	}

	return &Embedder{
		Config: cfg,
		client: cfg.newClient(),
	}, nil
}

func (e *Embedder) Embed(ctx context.Context, content string) (*provider.Embedding, error) {
	req := openai.EmbeddingRequest{
		Input: content,
		Model: openai.EmbeddingModel(e.model),
	}

	result, err := e.client.CreateEmbeddings(ctx, req)

	if err != nil {
		return nil, convertError(err)
	}

	return &provider.Embedding{
		Data: result.Data[0].Embedding,

		Usage: &provider.Usage{
			InputTokens:  result.Usage.PromptTokens,
			OutputTokens: result.Usage.CompletionTokens,
		},
	}, nil
}
