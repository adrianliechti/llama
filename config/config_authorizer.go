package config

import (
	"errors"
	"strings"

	"github.com/adrianliechti/llama/pkg/authorizer"
	"github.com/adrianliechti/llama/pkg/authorizer/oidc"
	"github.com/adrianliechti/llama/pkg/authorizer/static"
)

func (c *Config) registerAuthorizer(f *configFile) error {
	for _, a := range f.Authorizers {
		authorizer, err := createAuthorizer(a)

		if err != nil {
			return err
		}

		c.Authorizers = append(c.Authorizers, authorizer)
	}

	return nil
}

func createAuthorizer(cfg authorizerConfig) (authorizer.Provider, error) {
	switch strings.ToLower(cfg.Type) {
	case "static":
		return staticAuthorizer(cfg)

	case "oidc":
		return oidcAuthorizer(cfg)

	default:
		return nil, errors.New("invalid authorizer type: " + cfg.Type)
	}
}

func staticAuthorizer(cfg authorizerConfig) (authorizer.Provider, error) {
	return static.New(cfg.Token)
}

func oidcAuthorizer(cfg authorizerConfig) (authorizer.Provider, error) {
	return oidc.New(cfg.Issuer, cfg.Audience)
}
