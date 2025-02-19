package segmenter

import (
	"context"
	"errors"
)

type Provider interface {
	Segment(ctx context.Context, text string, options *SegmentOptions) ([]Segment, error)
}

var (
	ErrUnsupported = errors.New("unsupported type")
)

type SegmentOptions struct {
	FileName string

	SegmentLength  *int
	SegmentOverlap *int
}

type Segment struct {
	Text string
}
