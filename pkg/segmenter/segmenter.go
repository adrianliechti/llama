package segmenter

import (
	"context"
	"errors"
	"io"
)

type Provider interface {
	Segment(ctx context.Context, input File, options *SegmentOptions) ([]Segment, error)
}

var (
	ErrUnsupported = errors.New("unsupported type")
)

type SegmentOptions struct {
	SegmentLength  *int
	SegmentOverlap *int
}

type File struct {
	Name   string
	Reader io.Reader
}

type Segment struct {
	Name    string
	Content string
}
