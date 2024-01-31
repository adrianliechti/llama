package extracter

import (
	"context"
	"io"
)

type Provider interface {
	Extract(ctx context.Context, input File, options *ExtractOptions) (*Document, error)
}

type ExtractOptions struct {
}

type File struct {
	ID string

	Name    string
	Content io.Reader
}

type Document struct {
	Pages []Page
}

type Page struct {
	Blocks []Block
}

type Block struct {
	Text string
}
