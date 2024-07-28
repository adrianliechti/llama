package config

import (
	"errors"
	"strings"

	"github.com/adrianliechti/llama/pkg/provider"
	"github.com/adrianliechti/llama/pkg/provider/coqui"
	"github.com/adrianliechti/llama/pkg/provider/mimic"
	"github.com/adrianliechti/llama/pkg/provider/openai"
)

func (cfg *Config) RegisterSynthesizer(model string, s provider.Synthesizer) {
	cfg.RegisterModel(model)

	if cfg.synthesizer == nil {
		cfg.synthesizer = make(map[string]provider.Synthesizer)
	}

	cfg.synthesizer[model] = s
}

func createSynthesizer(cfg providerConfig, model string) (provider.Synthesizer, error) {
	switch strings.ToLower(cfg.Type) {

	case "coqui":
		return coquiSynthesizer(cfg, model)

	case "mimic":
		return mimicSynthesizer(cfg, model)

	case "openai":
		return openaiSynthesizer(cfg, model)

	default:
		return nil, errors.New("invalid synthesizer type: " + cfg.Type)
	}
}

func coquiSynthesizer(cfg providerConfig, model string) (provider.Synthesizer, error) {
	var options []coqui.Option

	return coqui.NewSynthesizer(cfg.URL, options...)
}

func mimicSynthesizer(cfg providerConfig, model string) (provider.Synthesizer, error) {
	var options []mimic.Option

	return mimic.NewSynthesizer(cfg.URL, options...)
}

func openaiSynthesizer(cfg providerConfig, model string) (provider.Synthesizer, error) {
	var options []openai.Option

	if cfg.URL != "" {
		options = append(options, openai.WithURL(cfg.URL))
	}

	if cfg.Token != "" {
		options = append(options, openai.WithToken(cfg.Token))
	}

	if model != "" {
		options = append(options, openai.WithModel(model))
	}

	return openai.NewSynthesizer(options...)
}
