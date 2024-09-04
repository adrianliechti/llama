package partitioner

import (
	"context"
	"errors"
	"io"
)

type Provider interface {
	Partition(ctx context.Context, input File, options *PartitionOptions) (*Document, error)
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

type Document struct {
	Name string

	Partitions []Partition
}

type Partition struct {
	ID      string
	Content string
}
