package config

import (
	"errors"
	"strings"

	"github.com/adrianliechti/llama/pkg/extracter"
	"github.com/adrianliechti/llama/pkg/extracter/tesseract"
	"github.com/adrianliechti/llama/pkg/extracter/unstructured"
)

func (c *Config) registerExtracters(f *configFile) error {
	for id, cfg := range f.Extracters {
		e, err := createExtracter(cfg)

		if err != nil {
			return err
		}

		c.extracters[id] = e
	}

	return nil
}

func createExtracter(cfg extracterConfig) (extracter.Provider, error) {
	switch strings.ToLower(cfg.Type) {
	case "tesseract":
		return tesseractExtracter(cfg)

	case "unstructured":
		return unstructuredExtracter(cfg)

	default:
		return nil, errors.New("invalid extracter type: " + cfg.Type)
	}
}

func tesseractExtracter(cfg extracterConfig) (extracter.Provider, error) {
	var options []tesseract.Option

	if cfg.URL != "" {
		options = append(options, tesseract.WithURL(cfg.URL))
	}

	return tesseract.New(options...)
}

func unstructuredExtracter(cfg extracterConfig) (extracter.Provider, error) {
	var options []unstructured.Option

	if cfg.URL != "" {
		options = append(options, unstructured.WithURL(cfg.URL))
	}

	return unstructured.New(options...)
}
