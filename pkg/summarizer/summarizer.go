package summarizer

import "context"

type Provider interface {
	Summarize(ctx context.Context, text string, options *SummarizerOptions) (*Summary, error)
}

type SummarizerOptions struct {
}

type Summary struct {
	Text string
}
