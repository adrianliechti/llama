package summarizer

import "context"

type Provider interface {
	Summarize(ctx context.Context, content string, options *SummarizerOptions) (*Result, error)
}

type SummarizerOptions struct {
}

type Result struct {
	Text     string
	Segments []string
}
