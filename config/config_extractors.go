package config

import (
	"errors"
	"strings"

	"github.com/adrianliechti/llama/pkg/extractor"
	"github.com/adrianliechti/llama/pkg/extractor/tesseract"
	"github.com/adrianliechti/llama/pkg/extractor/text"
	"github.com/adrianliechti/llama/pkg/extractor/unstructured"
)

func (cfg *Config) RegisterExtractor(model string, e extractor.Provider) {
	if cfg.extractors == nil {
		cfg.extractors = make(map[string]extractor.Provider)
	}

	cfg.extractors[model] = e
}

func (cfg *Config) registerExtractors(f *configFile) error {
	for id, c := range f.Extractors {
		extractor, err := createExtractor(c)

		if err != nil {
			return err
		}

		cfg.RegisterExtractor(id, extractor)
	}

	return nil
}

func createExtractor(cfg extractorConfig) (extractor.Provider, error) {
	switch strings.ToLower(cfg.Type) {
	case "text":
		return textExtractor(cfg)

	case "tesseract":
		return tesseractExtractor(cfg)

	case "unstructured":
		return unstructuredExtractor(cfg)

	default:
		return nil, errors.New("invalid extractor type: " + cfg.Type)
	}
}

func textExtractor(cfg extractorConfig) (extractor.Provider, error) {
	var options []text.Option

	return text.New(options...)
}

func tesseractExtractor(cfg extractorConfig) (extractor.Provider, error) {
	var options []tesseract.Option

	return tesseract.New(cfg.URL, options...)
}

func unstructuredExtractor(cfg extractorConfig) (extractor.Provider, error) {
	var options []unstructured.Option

	return unstructured.New(cfg.URL, options...)
}
