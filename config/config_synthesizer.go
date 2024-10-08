package config

import (
	"errors"
	"strings"

	"github.com/adrianliechti/llama/pkg/provider"
	"github.com/adrianliechti/llama/pkg/provider/elevenlabs"
	"github.com/adrianliechti/llama/pkg/provider/openai"
)

func (cfg *Config) RegisterSynthesizer(model string, p provider.Synthesizer) {
	cfg.RegisterModel(model)

	if cfg.synthesizer == nil {
		cfg.synthesizer = make(map[string]provider.Synthesizer)
	}

	cfg.synthesizer[model] = p
}

func (cfg *Config) Synthesizer(model string) (provider.Synthesizer, error) {
	if cfg.synthesizer != nil {
		if s, ok := cfg.synthesizer[model]; ok {
			return s, nil
		}
	}

	return nil, errors.New("synthesizer not found: " + model)
}

func createSynthesizer(cfg providerConfig, model modelContext) (provider.Synthesizer, error) {
	switch strings.ToLower(cfg.Type) {
	case "elevenlabs":
		return elevenlabsSynthesizer(cfg, model)

	case "openai":
		return openaiSynthesizer(cfg, model)

	default:
		return nil, errors.New("invalid synthesizer type: " + cfg.Type)
	}
}

func elevenlabsSynthesizer(cfg providerConfig, model modelContext) (provider.Synthesizer, error) {
	var options []elevenlabs.Option

	if cfg.Token != "" {
		options = append(options, elevenlabs.WithToken(cfg.Token))
	}

	return elevenlabs.NewSynthesizer(cfg.URL, model.ID, options...)
}

func openaiSynthesizer(cfg providerConfig, model modelContext) (provider.Synthesizer, error) {
	var options []openai.Option

	if cfg.Token != "" {
		options = append(options, openai.WithToken(cfg.Token))
	}

	return openai.NewSynthesizer(cfg.URL, model.ID, options...)
}
