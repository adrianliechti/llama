package static

import (
	"context"
	"errors"
	"net/http"
	"os"
	"strings"
)

type Provider struct {
	token string
}

func FromEnvironment() (*Provider, error) {
	token := os.Getenv("API_TOKEN")

	return New(token)
}

func New(token string) (*Provider, error) {
	return &Provider{
		token: token,
	}, nil
}

func (p *Provider) Verify(ctx context.Context, r *http.Request) error {
	if p.token == "" {
		return nil
	}

	header := r.Header.Get("Authorization")

	if header == "" {
		return errors.New("missing authorization header")
	}

	if !strings.HasPrefix(header, "Bearer ") {
		return errors.New("invalid authorization header")
	}

	token := strings.TrimPrefix(header, "Bearer ")

	if !strings.EqualFold(token, p.token) {
		return errors.New("invalid token")
	}

	return nil
}
