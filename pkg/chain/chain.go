package chain

import (
	"github.com/adrianliechti/wingman/pkg/provider"
)

type Provider interface {
	provider.Completer
}
