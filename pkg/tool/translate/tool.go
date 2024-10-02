package translate

import (
	"context"
	"errors"

	"github.com/adrianliechti/llama/pkg/tool"
	"github.com/adrianliechti/llama/pkg/translator"
)

var _ tool.Tool = &Tool{}

type Tool struct {
	name        string
	description string

	translator translator.Provider
}

func New(translator translator.Provider, options ...Option) (*Tool, error) {
	t := &Tool{
		name:        "translator",
		description: "Translate text to the given language.",

		translator: translator,
	}

	for _, option := range options {
		option(t)
	}

	if t.translator == nil {
		return nil, errors.New("missing translator provider")
	}

	return t, nil
}

func (t *Tool) Name() string {
	return t.name
}

func (t *Tool) Description() string {
	return t.description
}

func (*Tool) Parameters() any {
	return map[string]any{
		"type": "object",

		"properties": map[string]any{
			"text": map[string]any{
				"type":        "string",
				"description": "The text to translate",
			},

			"lang": map[string]any{
				"type":        "string",
				"description": "The target language code",
				"enum": []string{
					"de",
					"en",
					"fr",
					"it",
				},

				"default": "en",
			},
		},

		"required": []string{"query", "lang"},
	}
}

func (t *Tool) Execute(ctx context.Context, parameters map[string]any) (any, error) {
	text, ok := parameters["text"].(string)

	if !ok {
		return nil, errors.New("missing text parameter")
	}

	lang, ok := parameters["lang"].(string)

	if !ok {
		lang = "en"
	}

	options := &translator.TranslateOptions{
		Language: lang,
	}

	data, err := t.translator.Translate(ctx, text, options)

	if err != nil {
		return nil, err
	}

	return &Result{
		Language: lang,
		Text:     data.Content,
	}, nil
}
