package openai

import (
	"context"
	"errors"
	"time"

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

	var err error
	var result openai.EmbeddingResponse

	for i := 0; i < 20; i++ {
		result, err = c.client.CreateEmbeddings(ctx, req)

		if err != nil {
			e := &openai.APIError{}

			if errors.As(err, &e) {
				if e.Code == 429 {
					time.Sleep(2 * time.Second)
					continue
				}
			}

			return nil, convertError(err)
		}

		break
	}

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
