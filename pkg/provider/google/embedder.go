package google

import (
	"context"

	"github.com/adrianliechti/llama/pkg/provider"
	"github.com/google/generative-ai-go/genai"
)

var _ provider.Embedder = (*Embedder)(nil)

type Embedder struct {
	*Config
}

func NewEmbedder(model string, options ...Option) (*Embedder, error) {
	cfg := &Config{
		model: model,
	}

	for _, option := range options {
		option(cfg)
	}

	return &Embedder{
		Config: cfg,
	}, nil
}

func (e *Embedder) Embed(ctx context.Context, content string) (*provider.Embedding, error) {
	client, err := genai.NewClient(ctx, e.Options()...)

	if err != nil {
		return nil, err
	}

	defer client.Close()

	model := client.EmbeddingModel(e.model)

	resp, err := model.EmbedContent(ctx, genai.Text(content))

	if err != nil {
		return nil, convertError(err)
	}

	return &provider.Embedding{
		Data: resp.Embedding.Values,
	}, nil
}
