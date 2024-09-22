package adapter

import (
	"context"
	"strings"

	"github.com/adrianliechti/llama/pkg/provider"
	"github.com/adrianliechti/llama/pkg/summarizer"
	"github.com/adrianliechti/llama/pkg/text"
)

var _ summarizer.Provider = (*Adapter)(nil)

type Adapter struct {
	completer provider.Completer
}

func FromCompleter(completer provider.Completer) *Adapter {
	return &Adapter{
		completer: completer,
	}
}

func (a *Adapter) Summarize(ctx context.Context, content string, options *summarizer.SummarizerOptions) (*summarizer.Result, error) {
	splitter := text.NewSplitter()
	splitter.ChunkSize = 16000
	splitter.ChunkOverlap = 0

	var segments []string

	for _, part := range splitter.Split(content) {
		completion, err := a.completer.Complete(ctx, []provider.Message{
			{
				Role:    provider.MessageRoleUser,
				Content: "Write a concise summary of the following: \n" + part,
			},
		}, nil)

		if err != nil {
			return nil, err
		}

		segments = append(segments, completion.Message.Content)
	}

	completion, err := a.completer.Complete(ctx, []provider.Message{
		{
			Role:    provider.MessageRoleUser,
			Content: "Distill the following parts into a consolidated summary: \n" + strings.Join(segments, "\n\n"),
		},
	}, nil)

	if err != nil {
		return nil, err
	}

	result := &summarizer.Result{
		Text: completion.Message.Content,

		Segments: segments,
	}

	return result, nil
}
