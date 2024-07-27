package test

import (
	"context"

	"github.com/adrianliechti/llama/pkg/provider"
	"github.com/adrianliechti/llama/pkg/provider/ollama"
)

type TestContext struct {
	Context context.Context

	Embedder  provider.Embedder
	Completer provider.Completer
}

func NewContext() *TestContext {
	url := "http://localhost:11434"

	completer, _ := ollama.NewCompleter(url, ollama.WithModel("llama3.1:latest"))
	embedder, _ := ollama.NewEmbedder(url, ollama.WithModel("nomic-embed-text:latest"))

	return &TestContext{
		Context: context.Background(),

		Embedder:  embedder,
		Completer: completer,
	}
}
