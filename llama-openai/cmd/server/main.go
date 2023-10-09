package main

import (
	"chat/pkg/auth"
	"chat/pkg/auth/oidc"
	"chat/pkg/auth/static"
	"chat/pkg/dispatcher"
	"chat/pkg/server"
	"chat/provider"
	"chat/provider/codellama"
	"chat/provider/llama"
	"chat/provider/mistral"
	"chat/provider/openai"
)

func main() {
	var auth auth.Provider
	var providers []provider.Provider

	if p, err := static.FromEnvironment(); err == nil {
		auth = p
	}

	if p, err := oidc.FromEnvironment(); err == nil {
		auth = p
	}

	if p, err := openai.FromEnvironment(); err == nil {
		providers = append(providers, p)
	}

	if p, err := llama.FromEnvironment(); err == nil {
		providers = append(providers, p)
	}

	if p, err := codellama.FromEnvironment(); err == nil {
		providers = append(providers, p)
	}

	if p, err := mistral.FromEnvironment(); err == nil {
		providers = append(providers, p)
	}

	if auth == nil {
		panic("auth provider is not configured")
	}

	if len(providers) == 0 {
		panic("no provider configured")
	}

	provider, err := dispatcher.New(providers...)

	if err != nil {
		panic(err)
	}

	s := server.New(auth, provider)
	s.ListenAndServe(":8080")
}
