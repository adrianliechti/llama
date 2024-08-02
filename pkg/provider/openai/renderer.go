package openai

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"io"
	"strings"

	"github.com/adrianliechti/llama/pkg/provider"

	"github.com/google/uuid"
	"github.com/sashabaranov/go-openai"
)

var _ provider.Renderer = (*Renderer)(nil)

type Renderer struct {
	*Config
	client *openai.Client
}

func NewRenderer(options ...Option) (*Renderer, error) {
	cfg := &Config{
		model: string(openai.CreateImageModelDallE3),
	}

	for _, option := range options {
		option(cfg)
	}

	return &Renderer{
		Config: cfg,
		client: cfg.newClient(),
	}, nil
}

func (r *Renderer) Render(ctx context.Context, input string, options *provider.RenderOptions) (*provider.Image, error) {
	if options == nil {
		options = new(provider.RenderOptions)
	}

	req := openai.ImageRequest{
		Prompt: input,
		Model:  r.model,

		ResponseFormat: openai.CreateImageResponseFormatB64JSON,
	}

	result, err := r.client.CreateImage(ctx, req)

	if err != nil {
		return nil, convertError(err)
	}

	if len(result.Data) == 0 {
		return nil, errors.New("unable to render image")
	}

	data, err := base64.StdEncoding.DecodeString(result.Data[0].B64JSON)

	if err != nil {
		return nil, errors.New("invalid image data")
	}

	name := strings.ReplaceAll(uuid.NewString(), "-", "") + ".png"

	return &provider.Image{
		Name:    name,
		Content: io.NopCloser(bytes.NewReader(data)),
	}, nil
}
