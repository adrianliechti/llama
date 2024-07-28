package roundrobin

import (
	"context"
	"math/rand"

	"github.com/adrianliechti/llama/pkg/provider"
)

type Completer struct {
	completers []provider.Completer
}

func NewCompleter(completer ...provider.Completer) (provider.Completer, error) {
	c := &Completer{
		completers: completer,
	}

	return c, nil
}

func (c *Completer) Complete(ctx context.Context, messages []provider.Message, options *provider.CompleteOptions) (*provider.Completion, error) {
	index := rand.Intn(len(c.completers))
	provider := c.completers[index]

	return provider.Complete(ctx, messages, options)
}
