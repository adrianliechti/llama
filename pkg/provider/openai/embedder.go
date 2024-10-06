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
	req := openai.EmbeddingNewParams{
		Model:          openai.F(e.model),
		Input:          openai.F[openai.EmbeddingNewParamsInputUnion](shared.UnionString(content)),
		EncodingFormat: openai.F(openai.EmbeddingNewParamsEncodingFormatFloat),
	}

	result, err := e.embeddings.New(ctx, req)

	if err != nil {
		return nil, convertError(err)
	}

	var data []float32

	for _, v := range result.Data[0].Embedding {
		data = append(data, float32(v))
	}

	return &provider.Embedding{
		Data: data,

		Usage: &provider.Usage{
			InputTokens:  int(result.Usage.PromptTokens),
			OutputTokens: 0,
		},
	}, nil
}
