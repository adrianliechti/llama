package index

import (
	"context"
	"io"
	"strings"

	"github.com/adrianliechti/llama/pkg/extractor"
	"github.com/adrianliechti/llama/pkg/segmenter"
)

func (s *Handler) fileText(ctx context.Context, model string, name string, content io.Reader) (string, error) {
	p, err := s.Extractor(model)

	if err != nil {
		return "", err
	}

	input := extractor.File{
		Name:    name,
		Content: content,
	}

	document, err := p.Extract(ctx, input, &extractor.ExtractOptions{})

	if err != nil {
		return "", err
	}

	return document.Content, nil
}

func (s *Handler) textSegment(ctx context.Context, model, text string, segmentLength, segmentOverlap int) ([]string, error) {
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

	input := segmenter.File{
		Name:    "input.txt",
		Content: strings.NewReader(text),
	}

	segments, err := p.Segment(ctx, input, &segmenter.SegmentOptions{
		SegmentLength:  &segmentLength,
		SegmentOverlap: &segmentOverlap,
	})

	if err != nil {
		return nil, err
	}

	var result []string

	for _, s := range segments {
		result = append(result, s.Content)
	}

	return result, nil
}
