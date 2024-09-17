package draw

import (
	"context"
	"encoding/base64"
	"errors"
	"io"
	"net/http"

	"github.com/adrianliechti/llama/pkg/provider"
	"github.com/adrianliechti/llama/pkg/tool"
)

var _ tool.Tool = &Tool{}

type Tool struct {
	client   *http.Client
	renderer provider.Renderer
}

func New(options ...Option) (*Tool, error) {
	t := &Tool{
		client: http.DefaultClient,
	}

	for _, option := range options {
		option(t)
	}

	return t, nil
}

func (t *Tool) Name() string {
	return "draw"
}

func (t *Tool) Description() string {
	return "Draw images using stable diffusion based on a input prompt. Returns the image data as base64 encoded data"
}

func (*Tool) Parameters() any {
	return map[string]any{
		"type": "object",

		"properties": map[string]any{
			"prompt": map[string]any{
				"type":        "string",
				"description": "the prompt to create the image based from. must be in english - translate to english if needed.",
			},
		},

		"required": []string{"prompt"},
	}
}

func (t *Tool) Execute(ctx context.Context, parameters map[string]any) (any, error) {
	prompt, ok := parameters["prompt"].(string)

	if !ok {
		return nil, errors.New("missing prompt parameter")
	}

	options := &provider.RenderOptions{}

	image, err := t.renderer.Render(ctx, prompt, options)

	if err != nil {
		return nil, err
	}

	data, err := io.ReadAll(image.Content)

	if err != nil {
		return nil, err
	}

	return base64.StdEncoding.EncodeToString(data), nil
}
