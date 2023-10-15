package config

import (
	"errors"
	"os"

	"github.com/adrianliechti/llama/pkg/auth"
	"github.com/adrianliechti/llama/pkg/auth/oidc"
	"github.com/adrianliechti/llama/pkg/auth/static"

	"github.com/adrianliechti/llama/pkg/llm"
	"github.com/adrianliechti/llama/pkg/llm/codellama"
	"github.com/adrianliechti/llama/pkg/llm/llama"
	"github.com/adrianliechti/llama/pkg/llm/mistral"
	"github.com/adrianliechti/llama/pkg/llm/openai"
)

type Config struct {
	Addr string

	Auth auth.Provider
	LLM  llm.Provider
}

func FromEnvironment() (*Config, error) {
	addr := addrFromEnvironment()

	auth := authFromEnvironment()
	llm := llmFromEnvironment()

	if auth == nil {
		return nil, errors.New("no auth provider configured")
	}

	if llm == nil {
		return nil, errors.New("no llm provider configured")
	}

	return &Config{
		Addr: addr,

		Auth: auth,
		LLM:  llm,
	}, nil
}

func addrFromEnvironment() string {
	port := os.Getenv("PORT")

	if port == "" {
		port = "8080"
	}

	return ":" + port
}

func authFromEnvironment() auth.Provider {
	if p, err := oidc.FromEnvironment(); err == nil {
		return p
	}

	if p, err := static.FromEnvironment(); err == nil {
		return p
	}

	return nil
}

func llmFromEnvironment() llm.Provider {
	if p, err := openai.FromEnvironment(); err == nil {
		return p
	}

	if p, err := llama.FromEnvironment(); err == nil {
		return p
	}

	if p, err := codellama.FromEnvironment(); err == nil {
		return p
	}

	if p, err := mistral.FromEnvironment(); err == nil {
		return p
	}

	return nil
}
