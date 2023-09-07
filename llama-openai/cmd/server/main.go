package main

import (
	"chat/pkg/auth"
	"chat/pkg/auth/oidc"
	"chat/pkg/auth/static"
	"chat/pkg/server"
	"chat/provider"
	"chat/provider/llama"
)

func main() {
	var auth auth.Provider
	var provider provider.Provider

	if p, err := static.FromEnvironment(); err == nil {
		auth = p
	}

	if p, err := oidc.FromEnvironment(); err == nil {
		auth = p
	}

	if p, err := llama.FromEnvironment(); err == nil {
		provider = p
	}

	if auth == nil {
		panic("auth provider is not configured")
	}

	if provider == nil {
		panic("provider is not configured")
	}

	s := server.New(auth, provider)
	s.ListenAndServe(":8080")
}
