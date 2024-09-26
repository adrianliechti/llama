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
	client      *http.Client
	synthesizer provider.Synthesizer
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
	return "speak"
}

func (t *Tool) Description() string {
	return "Synthesize speech from text using a TTS (text-to-speech) model on a input prompt. Returns a URL to the generated audio file. Render the URL as markdown ```[Download Audio](url)```"
}

func (*Tool) Parameters() any {
	return map[string]any{
		"type": "object",

		"properties": map[string]any{
			"prompt": map[string]any{
				"type":        "string",
				"description": "the prompt to create the audio file based from. can be in orgiginal language and should not be translated.",
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
