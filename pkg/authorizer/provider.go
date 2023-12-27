package authorizer

import (
	"context"
	"net/http"
)

type Provider interface {
	Verify(ctx context.Context, r *http.Request) error
}
