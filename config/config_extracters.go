package config

import (
	"errors"
	"strings"

	"github.com/adrianliechti/llama/pkg/extracter"
	"github.com/adrianliechti/llama/pkg/extracter/tesseract"
	"github.com/adrianliechti/llama/pkg/extracter/text"
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
	case "text":
		return textExtracter(cfg)

	case "tesseract":
		return tesseractExtracter(cfg)

	case "unstructured":
		return unstructuredExtracter(cfg)

	default:
		return nil, errors.New("invalid extracter type: " + cfg.Type)
	}
}

func textExtracter(cfg extracterConfig) (extracter.Provider, error) {
	var options []text.Option

	return text.New(options...)
}

func tesseractExtracter(cfg extracterConfig) (extracter.Provider, error) {
	var options []tesseract.Option

	return tesseract.New(cfg.URL, options...)
}

func unstructuredExtracter(cfg extracterConfig) (extracter.Provider, error) {
	var options []unstructured.Option

	return unstructured.New(cfg.URL, options...)
}
