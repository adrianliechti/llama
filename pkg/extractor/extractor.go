package extractor

import (
	"context"
	"errors"
	"io"
)

type Provider interface {
	Extract(ctx context.Context, input File, options *ExtractOptions) (*Document, error)
}

var (
	ErrUnsupported = errors.New("unsupported type")
)

type ExtractOptions struct {
}

type File struct {
	ID string

	Name string

	URL     string
	Content io.Reader
}

type Document struct {
	ID string

	Name    string
	Content string
}
