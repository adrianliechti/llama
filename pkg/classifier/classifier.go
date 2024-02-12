package classifier

import (
	"context"
)

type Provider interface {
	Classify(ctx context.Context, input string) (string, error)
}

type Class struct {
	Name string

	Description string
}
