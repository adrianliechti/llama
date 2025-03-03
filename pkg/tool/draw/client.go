package draw

import (
	"context"
	"errors"
	"io"
	"net/url"
	"os"
	"path/filepath"

	"github.com/adrianliechti/wingman/pkg/provider"
	"github.com/adrianliechti/wingman/pkg/tool"
	"github.com/google/uuid"
)

var _ tool.Provider = (*Client)(nil)

type Client struct {
	provider provider.Renderer
}

func New(provider provider.Renderer, options ...Option) (*Client, error) {
	c := &Client{
		provider: provider,
	}

	for _, option := range options {
		option(c)
	}

	return c, nil
}

func (c *Client) Tools(ctx context.Context) ([]tool.Tool, error) {
	return []tool.Tool{
		{
			Name:        "draw_image",
			Description: "Generate images based based on user-provided prompts. Returns a URL to download the generated image. Editing images is not supported.",

			Parameters: map[string]any{
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
			},
		},
	}, nil
}

func (c *Client) Execute(ctx context.Context, name string, parameters map[string]any) (any, error) {
	if name != "draw_image" {
		return nil, tool.ErrInvalidTool
	}

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

	image, err := c.provider.Render(ctx, prompt, options)

	if err != nil {
		return nil, err
	}

	path := uuid.New().String() + ".png"

	os.MkdirAll(filepath.Join("public", "files"), 0755)

	f, err := os.Create(filepath.Join("public", "files", path))

	if err != nil {
		return nil, err
	}

	if _, err := io.Copy(f, image.Reader); err != nil {
		return nil, err
	}

	url, err := url.JoinPath(os.Getenv("BASE_URL"), "files/"+path)

	if err != nil {
		return nil, err
	}

	return Result{
		URL: url,

		//Style:  string(options.Style),
		//Prompt: prompt,
	}, nil
}
