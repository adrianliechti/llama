package limiter

import (
	"context"

	"github.com/adrianliechti/llama/pkg/provider"

	"golang.org/x/time/rate"
)

type Transcriber interface {
	Limiter
	provider.Transcriber
}

type limitedTranscriber struct {
	limiter  *rate.Limiter
	provider provider.Transcriber
}

func NewTranscriber(l *rate.Limiter, p provider.Transcriber) Transcriber {
	return &limitedTranscriber{
		limiter:  l,
		provider: p,
	}
}

func (p *limitedTranscriber) limiterSetup() {
}

func (p *limitedTranscriber) Transcribe(ctx context.Context, input provider.File, options *provider.TranscribeOptions) (*provider.Transcription, error) {
	if p.limiter != nil {
		p.limiter.Wait(ctx)
	}

	return p.provider.Transcribe(ctx, input, options)
}
