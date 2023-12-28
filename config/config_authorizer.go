package config

import (
	"errors"
	"strings"

	"github.com/adrianliechti/llama/pkg/authorizer"
	"github.com/adrianliechti/llama/pkg/authorizer/oidc"
	"github.com/adrianliechti/llama/pkg/authorizer/static"
)

func createAuthorizer(c authorizerConfig) (authorizer.Provider, error) {
	switch strings.ToLower(c.Type) {
	case "static":
		return staticAuthorizer(c)

	case "oidc":
		return oidcAuthorizer(c)

	default:
		return nil, errors.New("invalid authorizer type: " + c.Type)
	}
}

func staticAuthorizer(c authorizerConfig) (authorizer.Provider, error) {
	return static.New(c.Token)
}

func oidcAuthorizer(c authorizerConfig) (authorizer.Provider, error) {
	return oidc.New(c.Issuer, c.Audience)
}
