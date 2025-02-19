package provider

import (
	"context"
	"io"
)

type Synthesizer interface {
	Synthesize(ctx context.Context, input string, options *SynthesizeOptions) (*Synthesis, error)
}

type SynthesizeOptions struct {
	Voice string
}

type Synthesis struct {
	ID string

	Name   string
	Reader io.ReadCloser
}
