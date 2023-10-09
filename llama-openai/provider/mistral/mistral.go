package mistral

import (
	"errors"
	"os"

	"chat/provider/llama"
)

func FromEnvironment() (*llama.Provider, error) {
	url := os.Getenv("MISTRAL_URL")

	if url == "" {
		return nil, errors.New("MISTRAL_URL is not set")
	}

	model := os.Getenv("MISTRAL_MODEL")

	if model == "" {
		model = "default"
	}

	return llama.New(url, model, "")
}
