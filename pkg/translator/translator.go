package translator

import (
	"context"
)

type Provider interface {
	Translate(ctx context.Context, content string, options *TranslateOptions) (*Translation, error)
}

type TranslateOptions struct {
	Language string
}

type Translation struct {
	ID string

	Content string
}
