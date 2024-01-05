package classifier

import (
	"context"
)

type Provider interface {
	Categorize(ctx context.Context, input string) (string, error)
}
