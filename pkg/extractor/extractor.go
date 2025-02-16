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
	Name string

	URL    string
	Reader io.Reader
}

type Document struct {
	Name string

	Content     string
	ContentType string
}
