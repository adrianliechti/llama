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

	Name    string
	Content io.Reader
}

type Document struct {
	Name string

	Blocks []Block
}

type Block struct {
	ID      string
	Content string
}
