package crawler

import (
	"context"
	"errors"

	"github.com/adrianliechti/llama/pkg/extractor"
	"github.com/adrianliechti/llama/pkg/tool"
)

var _ tool.Tool = &Tool{}

type Tool struct {
	name        string
	description string

	extractor extractor.Provider
}

func New(extractor extractor.Provider, options ...Option) (*Tool, error) {
	t := &Tool{
		name:        "crawler",
		description: "fetch and return the markdown content from a given URL, including website pages, YouTube video transcriptions, and similar sources",

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
	return t.name
}

func (t *Tool) Description() string {
	return t.description
}

func (*Tool) Parameters() map[string]any {
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
