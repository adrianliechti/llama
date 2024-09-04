package chain

import (
	"github.com/adrianliechti/llama/pkg/provider"
)

type Provider interface {
	provider.Completer
}
