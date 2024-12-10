package speak

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

	client      *http.Client
	synthesizer provider.Synthesizer
}

func New(options ...Option) (*Tool, error) {
	t := &Tool{
		name:        "speak",
		description: "Synthesize speech from text using a TTS (text-to-speech) model on a input prompt. Returns a URL to the generated audio file. Render the URL as markdown ```[prompt](url)```",

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
				"description": "text to generate audio for in orgiginal language",
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

	options := &provider.SynthesizeOptions{}

	synthesis, err := t.synthesizer.Synthesize(ctx, prompt, options)

	if err != nil {
		return nil, err
	}

	name := uuid.New().String() + ".wav"

	os.MkdirAll(filepath.Join("public", "files"), 0755)

	f, err := os.Create(filepath.Join("public", "files", name))

	if err != nil {
		return nil, err
	}

	if _, err := io.Copy(f, synthesis.Content); err != nil {
		return nil, err
	}

	url, err := url.JoinPath(os.Getenv("BASE_URL"), "files/"+name)

	if err != nil {
		return nil, err
	}

	return Result{
		URL: url,
	}, nil
}
