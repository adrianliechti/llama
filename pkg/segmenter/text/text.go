package text

import (
	"context"
	"fmt"
	"io"
	"path"
	"strings"

	"github.com/adrianliechti/llama/pkg/segmenter"
	"github.com/adrianliechti/llama/pkg/text"
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

func (p *Provider) Segment(ctx context.Context, input segmenter.File, options *segmenter.SegmentOptions) ([]segmenter.Segment, error) {
	if options == nil {
		options = new(segmenter.SegmentOptions)
	}

	data, err := io.ReadAll(input.Content)

	if err != nil {
		return nil, err
	}

	splitter := text.NewSplitter()

	if options.SegmentLength != nil {
		splitter.ChunkSize = *options.SegmentLength
	}

	if options.SegmentOverlap != nil {
		splitter.ChunkOverlap = *options.SegmentOverlap
	}

	if sep := getSeperators(input); len(sep) > 0 {
		splitter.Separators = sep
	}

	var segments []segmenter.Segment

	for i, chunk := range splitter.Split(string(data)) {
		segment := segmenter.Segment{
			Name:    fmt.Sprintf("%s#%d", input.Name, i),
			Content: chunk,
		}

		segments = append(segments, segment)
	}

	return segments, nil
}

func getSeperators(input segmenter.File) []string {
	switch strings.ToLower(path.Ext(input.Name)) {
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
