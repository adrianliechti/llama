package text

import (
	"context"
	"path"
	"strings"

	"github.com/adrianliechti/wingman/pkg/segmenter"
	"github.com/adrianliechti/wingman/pkg/text"
)

var _ segmenter.Provider = &Provider{}

type Provider struct {
}

type Option func(*Provider)

func New(options ...Option) (*Provider, error) {
	p := &Provider{}

	for _, option := range options {
		option(p)
	}

	return p, nil
}

func (p *Provider) Segment(ctx context.Context, input string, options *segmenter.SegmentOptions) ([]segmenter.Segment, error) {
	if options == nil {
		options = new(segmenter.SegmentOptions)
	}

	splitter := text.NewSplitter()

	if options.SegmentLength != nil {
		splitter.ChunkSize = *options.SegmentLength
	}

	if options.SegmentOverlap != nil {
		splitter.ChunkOverlap = *options.SegmentOverlap
	}

	if sep := getSeperators(options.FileName); len(sep) > 0 {
		splitter.Separators = sep
	}

	var segments []segmenter.Segment

	for _, chunk := range splitter.Split(input) {
		segment := segmenter.Segment{
			Text: chunk,
		}

		segments = append(segments, segment)
	}

	return segments, nil
}

func getSeperators(name string) []string {
	switch strings.ToLower(path.Ext(name)) {
	case ".cs":
		return languageCSharp
	case ".cpp":
		return languageCPP
	case ".go":
		return languageGo
	case ".java":
		return languageJava
	case ".kt":
		return languageKotlin
	case ".js", ".jsm":
		return languageJavaScript
	case ".ts", ".tsx":
		return languageTypeScript
	case ".py":
		return languagePython
	case ".rb":
		return languageRuby
	case ".rs":
		return languageRust
	case ".sc", ".scala":
		return languageScala
	case ".swift":
		return languageSwift
	}

	return nil
}
