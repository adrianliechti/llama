package main

import (
	"os"

	"chat/pkg/server"

	"chat/pkg/auth"
	"chat/pkg/auth/oidc"
	"chat/pkg/auth/static"

	"chat/provider"
	"chat/provider/codellama"
	"chat/provider/llama"
	"chat/provider/mistral"
	"chat/provider/openai"
)

func main() {
	var auth auth.Provider
	var provider provider.Provider

	port := os.Getenv("PORT")

	if port == "" {
		port = "8080"
	}

	if p, err := static.FromEnvironment(); err == nil {
		auth = p
	}

	if p, err := oidc.FromEnvironment(); err == nil {
		auth = p
	}

	if p, err := openai.FromEnvironment(); err == nil {
		provider = p
	}

	if p, err := llama.FromEnvironment(); err == nil {
		provider = p
	}

	if p, err := codellama.FromEnvironment(); err == nil {
		provider = p
	}

	if p, err := mistral.FromEnvironment(); err == nil {
		provider = p
	}

	if auth == nil {
		panic("auth provider is not configured")
	}

	if provider == nil {
		panic("no provider configured")
	}

	s := server.New(auth, provider)
	s.ListenAndServe(":" + port)
}
