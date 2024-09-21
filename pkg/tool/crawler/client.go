package crawler

import (
	"context"
	"errors"
	"net/http"

	"github.com/adrianliechti/llama/pkg/extractor"
	"github.com/adrianliechti/llama/pkg/tool"
)

var _ tool.Tool = &Tool{}

type Tool struct {
	client *http.Client

	extractor extractor.Provider
}

func New(extractor extractor.Provider, options ...Option) (*Tool, error) {
	t := &Tool{
		client: http.DefaultClient,

		extractor: extractor,
	}

	for _, option := range options {
		option(t)
	}

	if t.extractor == nil {
		return nil, errors.New("missing extractor provider")
	}

	return t, nil
}

func (t *Tool) Name() string {
	return "crawler"
}

func (t *Tool) Description() string {
	return "return the content of a website as markdown"
}

func (*Tool) Parameters() any {
	return map[string]any{
		"type": "object",

		"properties": map[string]any{
			"url": map[string]any{
				"type":        "string",
				"description": "the URL of the website to crawl staring with http:// or https://",
			},
		},

		"required": []string{"url"},
	}
}

func (t *Tool) Execute(ctx context.Context, parameters map[string]any) (any, error) {
	url, ok := parameters["url"].(string)

	if !ok {
		return nil, errors.New("missing url parameter")
	}

	input := extractor.File{
		URL: url,
	}

	options := &extractor.ExtractOptions{}

	document, err := t.extractor.Extract(ctx, input, options)

	if err != nil {
		return nil, err
	}

	return document.Content, nil
}