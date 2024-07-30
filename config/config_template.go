package config

import (
	"errors"
	"os"

	"github.com/adrianliechti/llama/pkg/prompt"
)

func parseTemplate(val string) (*prompt.Template, error) {
	if val == "" {
		return nil, errors.New("empty prompt")
	}

	if data, err := os.ReadFile(val); err == nil {
		return prompt.NewTemplate(string(data))
	}

	return prompt.NewTemplate(val)
}
