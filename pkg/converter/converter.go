package converter

import (
	"context"
	"errors"
	"io"
)

type Provider interface {
	Convert(ctx context.Context, input File, options *ConvertOptions) (*Document, error)
}

var (
	ErrUnsupported = errors.New("unsupported type")
)

type ConvertOptions struct {
}

type File struct {
	ID string

	Name    string
	Content io.Reader
}

type Document struct {
	ID string

	Content string
}
