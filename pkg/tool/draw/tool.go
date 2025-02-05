package draw

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	"github.com/adrianliechti/llama/pkg/provider"
	"github.com/adrianliechti/llama/pkg/tool"
	"github.com/google/uuid"
)

var _ tool.Tool = &Tool{}

type Tool struct {
	name        string
	description string

	client   *http.Client
	renderer provider.Renderer
}

func New(options ...Option) (*Tool, error) {
	t := &Tool{
		name:        "draw",
		description: "Generate images based based on user-provided prompts. Returns a URL to download the generated image. Editing images is not supported.",

		client: http.DefaultClient,
	}

	for _, option := range options {
		option(t)
	}

	return t, nil
}

func (t *Tool) Name() string {
	return t.name
}

func (t *Tool) Description() string {
	return t.description
}

func (*Tool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",

		"properties": map[string]any{
			"prompt": map[string]any{
				"type":        "string",
				"description": "detailed text description of the image to generate. must be english.",
			},

			"style": map[string]any{
				"type":        "string",
				"description": "style of the image. defaults to vivid.",

				"enum": []string{
					"vivid",
					"natural",
				},
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

	options := &provider.RenderOptions{
		Style: provider.ImageStyleVivid,
	}

	if style, ok := parameters["style"].(string); ok {
		if style == "vivid" {
			options.Style = provider.ImageStyleVivid
		}

		if style == "natural" {
			options.Style = provider.ImageStyleNatural
		}
	}

	image, err := t.renderer.Render(ctx, prompt, options)

	if err != nil {
		return nil, err
	}

	name := uuid.New().String() + ".png"

	os.MkdirAll(filepath.Join("public", "files"), 0755)

	f, err := os.Create(filepath.Join("public", "files", name))

	if err != nil {
		return nil, err
	}

	if _, err := io.Copy(f, image.Content); err != nil {
		return nil, err
	}

	url, err := url.JoinPath(os.Getenv("BASE_URL"), "files/"+name)

	if err != nil {
		return nil, err
	}

	return Result{
		URL: url,

		//Style:  string(options.Style),
		//Prompt: prompt,
	}, nil
}
