package index

import (
	"context"
	"io"

	"github.com/adrianliechti/wingman/pkg/extractor"
	"github.com/adrianliechti/wingman/pkg/segmenter"
)

func (s *Handler) readText(ctx context.Context, model string, name string, content io.Reader) (string, error) {
	p, err := s.Extractor(model)

	if err != nil {
		return "", err
	}

	input := extractor.File{
		Name:   name,
		Reader: content,
	}

	document, err := p.Extract(ctx, input, &extractor.ExtractOptions{})

	if err != nil {
		return "", err
	}

	return document.Content, nil
}

func (s *Handler) segmentText(ctx context.Context, model, text string, segmentLength, segmentOverlap int) ([]string, error) {
	if segmentLength <= 0 {
		segmentLength = 1500
	}

	if segmentOverlap <= 0 {
		segmentOverlap = 0
	}

	p, err := s.Segmenter(model)

	if err != nil {
		return nil, err
	}

	segments, err := p.Segment(ctx, text, &segmenter.SegmentOptions{
		SegmentLength:  &segmentLength,
		SegmentOverlap: &segmentOverlap,
	})

	if err != nil {
		return nil, err
	}

	var result []string

	for _, s := range segments {
		result = append(result, s.Text)
	}

	return result, nil
}
