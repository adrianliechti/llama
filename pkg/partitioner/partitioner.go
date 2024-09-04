package partitioner

import (
	"context"
	"errors"
	"io"
)

type Provider interface {
	Partition(ctx context.Context, input File, options *PartitionOptions) ([]Partition, error)
}

var (
	ErrUnsupported = errors.New("unsupported type")
)

type PartitionOptions struct {
}

type File struct {
	ID string

	Name    string
	Content io.Reader
}

type Partition struct {
	ID      string
	Content string
}
