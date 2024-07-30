package config

import (
	"errors"
	"strings"

	"github.com/adrianliechti/llama/pkg/otel"
	"github.com/adrianliechti/llama/pkg/provider"
	"github.com/adrianliechti/llama/pkg/provider/groq"
	"github.com/adrianliechti/llama/pkg/provider/openai"
	"github.com/adrianliechti/llama/pkg/provider/whisper"
)

func (cfg *Config) RegisterTranscriber(name, model string, p provider.Transcriber) {
	cfg.RegisterModel(model)

	if cfg.transcriber == nil {
		cfg.transcriber = make(map[string]provider.Transcriber)
	}

	transcriber, ok := p.(otel.ObservableTranscriber)

	if !ok {
		transcriber = otel.NewTranscriber(name, model, p)
	}

	cfg.transcriber[model] = transcriber
}

func createTranscriber(cfg providerConfig, model string) (provider.Transcriber, error) {
	switch strings.ToLower(cfg.Type) {

	case "groq":
		return groqTranscriber(cfg, model)

	case "openai":
		return openaiTranscriber(cfg, model)

	case "whisper":
		return whisperTranscriber(cfg, model)

	default:
		return nil, errors.New("invalid transcriber type: " + cfg.Type)
	}
}

func groqTranscriber(cfg providerConfig, model string) (provider.Transcriber, error) {
	var options []groq.Option

	if cfg.Token != "" {
		options = append(options, groq.WithToken(cfg.Token))
	}

	if model != "" {
		options = append(options, groq.WithModel(model))
	}

	return groq.NewTranscriber(options...)
}

func openaiTranscriber(cfg providerConfig, model string) (provider.Transcriber, error) {
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

	return openai.NewTranscriber(options...)
}

func whisperTranscriber(cfg providerConfig, model string) (provider.Transcriber, error) {
	var options []whisper.Option

	return whisper.NewTranscriber(cfg.URL, options...)
}
