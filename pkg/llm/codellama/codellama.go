package codellama

import (
	"errors"
	"os"

	"github.com/adrianliechti/llama/pkg/llm/llama"
)

func FromEnvironment() (*llama.Provider, error) {
	url := os.Getenv("CODELLAMA_URL")

	if url == "" {
		return nil, errors.New("CODELLAMA_URL is not set")
	}

	model := os.Getenv("CODELLAMA_MODEL")

	if model == "" {
		model = "default"
	}

	system := os.Getenv("CODELLAMA_SYSTEM")

	return llama.New(url, model, system)
}
