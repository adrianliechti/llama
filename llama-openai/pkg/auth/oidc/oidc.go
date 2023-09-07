package oidc

import (
	"context"
	"errors"
	"net/http"
	"os"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"
)

type Provider struct {
	provider *oidc.Provider
	verifier *oidc.IDTokenVerifier
}

func FromEnvironment() (*Provider, error) {
	issuer := os.Getenv("OIDC_ISSUER")
	audience := os.Getenv("OIDC_AUDIENCE")

	if issuer == "" {
		return nil, errors.New("missing OIDC_ISSUER")
	}

	if audience == "" {
		return nil, errors.New("missing OIDC_AUDIENCE")
	}

	return New(issuer, audience)
}

func New(issuer, audience string) (*Provider, error) {
	cfg := &oidc.Config{
		ClientID: audience,
	}

	provider, err := oidc.NewProvider(context.Background(), issuer)

	if err != nil {
		return nil, err
	}

	verifier := provider.Verifier(cfg)

	return &Provider{
		provider: provider,
		verifier: verifier,
	}, nil
}

func (p *Provider) Verify(ctx context.Context, r *http.Request) error {
	header := r.Header.Get("Authorization")

	if header == "" {
		return errors.New("missing authorization header")
	}

	if !strings.HasPrefix(header, "Bearer ") {
		return errors.New("invalid authorization header")
	}

	token := strings.TrimPrefix(header, "Bearer ")

	idtoken, err := p.verifier.Verify(ctx, token)

	if err != nil {
		return err
	}

	_ = idtoken
	return nil
}
