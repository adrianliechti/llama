package openai

import (
	"context"

	"github.com/adrianliechti/llama/pkg/provider"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/shared"
)

var _ provider.Embedder = (*Embedder)(nil)

type Embedder struct {
	*Config
	embeddings *openai.EmbeddingService
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
		Config:     cfg,
		embeddings: openai.NewEmbeddingService(cfg.Options()...),
	}, nil
}

func (e *Embedder) Embed(ctx context.Context, content string) (*provider.Embedding, error) {
	result, err := e.embeddings.New(ctx, openai.EmbeddingNewParams{
		Model:          openai.F(e.model),
		Input:          openai.F[openai.EmbeddingNewParamsInputUnion](shared.UnionString(content)),
		EncodingFormat: openai.F(openai.EmbeddingNewParamsEncodingFormatFloat),
	})

	if err != nil {
		return nil, convertError(err)
	}

	return &provider.Embedding{
		Data: toFloat32(result.Data[0].Embedding),

		Usage: &provider.Usage{
			InputTokens:  int(result.Usage.PromptTokens),
			OutputTokens: 0,
		},
	}, nil
}

func toFloat32(input []float64) []float32 {
	result := make([]float32, len(input))

	for i, v := range input {
		result[i] = float32(v)
	}

	return result
}
