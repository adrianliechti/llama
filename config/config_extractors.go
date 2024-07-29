package config

import (
	"errors"
	"strings"

	"github.com/adrianliechti/llama/pkg/extractor"
	"github.com/adrianliechti/llama/pkg/extractor/code"
	"github.com/adrianliechti/llama/pkg/extractor/tesseract"
	"github.com/adrianliechti/llama/pkg/extractor/text"
	"github.com/adrianliechti/llama/pkg/extractor/tika"
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

	case "code":
		return codeExtractor(cfg)

	case "tesseract":
		return tesseractExtractor(cfg)

	case "tika":
		return tikaExtractor(cfg)

	case "unstructured":
		return unstructuredExtractor(cfg)

	default:
		return nil, errors.New("invalid extractor type: " + cfg.Type)
	}
}

func textExtractor(cfg extractorConfig) (extractor.Provider, error) {
	var options []text.Option

	if cfg.ChunkSize != nil {
		options = append(options, text.WithChunkSize(*cfg.ChunkSize))
	}

	if cfg.ChunkOverlap != nil {
		options = append(options, text.WithChunkOverlap(*cfg.ChunkOverlap))
	}

	return text.New(options...)
}

func codeExtractor(cfg extractorConfig) (extractor.Provider, error) {
	var options []code.Option

	if cfg.ChunkSize != nil {
		options = append(options, code.WithChunkSize(*cfg.ChunkSize))
	}

	if cfg.ChunkOverlap != nil {
		options = append(options, code.WithChunkOverlap(*cfg.ChunkOverlap))
	}

	return code.New(options...)
}

func tesseractExtractor(cfg extractorConfig) (extractor.Provider, error) {
	var options []tesseract.Option

	if cfg.ChunkSize != nil {
		options = append(options, tesseract.WithChunkSize(*cfg.ChunkSize))
	}

	if cfg.ChunkOverlap != nil {
		options = append(options, tesseract.WithChunkOverlap(*cfg.ChunkOverlap))
	}

	return tesseract.New(cfg.URL, options...)
}

func tikaExtractor(cfg extractorConfig) (extractor.Provider, error) {
	var options []tika.Option

	if cfg.ChunkSize != nil {
		options = append(options, tika.WithChunkSize(*cfg.ChunkSize))
	}

	if cfg.ChunkOverlap != nil {
		options = append(options, tika.WithChunkOverlap(*cfg.ChunkOverlap))
	}

	return tika.New(cfg.URL, options...)
}

func unstructuredExtractor(cfg extractorConfig) (extractor.Provider, error) {
	var options []unstructured.Option

	if cfg.ChunkSize != nil {
		options = append(options, unstructured.WithChunkSize(*cfg.ChunkSize))
	}

	if cfg.ChunkOverlap != nil {
		options = append(options, unstructured.WithChunkOverlap(*cfg.ChunkOverlap))
	}

	return unstructured.New(cfg.URL, options...)
}
