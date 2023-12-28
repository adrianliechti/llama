package config

import (
	"errors"
	"strings"

	"github.com/adrianliechti/llama/pkg/authorizer"
	"github.com/adrianliechti/llama/pkg/authorizer/oidc"
	"github.com/adrianliechti/llama/pkg/authorizer/static"
)

func createAuthorizer(a authorizerConfig) (authorizer.Provider, error) {
	switch strings.ToLower(a.Type) {
	case "static":
		return staticAuthorizer(a)

	case "oidc":
		return oidcAuthorizer(a)

	default:
		return nil, errors.New("invalid authorizer type: " + a.Type)
	}
}

func staticAuthorizer(a authorizerConfig) (authorizer.Provider, error) {
	return static.New(a.Token)
}

func oidcAuthorizer(a authorizerConfig) (authorizer.Provider, error) {
	return oidc.New(a.Issuer, a.Audience)
}
