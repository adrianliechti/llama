package rag

import (
	"context"
	"errors"

	"github.com/adrianliechti/llama/pkg/provider"
)

var (
	_ provider.Completer = &Provider{}
)

type Provider struct {
}

func (*Provider) Complete(ctx context.Context, messages []provider.Message, options *provider.CompleteOptions) (*provider.Completion, error) {
	return nil, errors.ErrUnsupported
}
