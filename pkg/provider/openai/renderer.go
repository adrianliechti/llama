package openai

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"io"

	"github.com/adrianliechti/llama/pkg/provider"
	"github.com/google/uuid"

	"github.com/openai/openai-go"
)

var _ provider.Renderer = (*Renderer)(nil)

type Renderer struct {
	*Config
	images *openai.ImageService
}

func NewRenderer(url, model string, options ...Option) (*Renderer, error) {
	cfg := &Config{
		url:   url,
		model: model,
	}

	for _, option := range options {
		option(cfg)
	}

	return &Renderer{
		Config: cfg,
		images: openai.NewImageService(cfg.Options()...),
	}, nil
}

func (r *Renderer) Render(ctx context.Context, input string, options *provider.RenderOptions) (*provider.Image, error) {
	if options == nil {
		options = new(provider.RenderOptions)
	}

	image, err := r.images.Generate(ctx, openai.ImageGenerateParams{
		Model:  openai.F(r.model),
		Prompt: openai.F(input),

		ResponseFormat: openai.F(openai.ImageGenerateParamsResponseFormatB64JSON),
	})

	if err != nil {
		return nil, convertError(err)
	}

	data, err := base64.StdEncoding.DecodeString(image.Data[0].B64JSON)

	if err != nil {
		return nil, errors.New("invalid image data")
	}

	id := uuid.NewString()

	return &provider.Image{
		ID:   id,
		Name: id + ".png",

		Content: io.NopCloser(bytes.NewReader(data)),
	}, nil
}
