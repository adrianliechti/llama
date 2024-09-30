package config

import (
	"errors"
	"os"

	"github.com/adrianliechti/llama/pkg/template"
)

func parseTemplate(val string) (*template.Template, error) {
	if val == "" {
		return nil, errors.New("empty template")
	}

	if data, err := os.ReadFile(val); err == nil {
		return template.NewTemplate(string(data))
	}

	return template.NewTemplate(val)
}
