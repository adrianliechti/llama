package config

import (
	"errors"
	"os"

	"github.com/adrianliechti/wingman/pkg/template"
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
