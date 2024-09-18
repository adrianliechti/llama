package config

import (
	"errors"

	"github.com/adrianliechti/llama/pkg/provider"

	"golang.org/x/time/rate"
	"gopkg.in/yaml.v3"
)

func (cfg *Config) registerProviders(f *configFile) error {
	for _, p := range f.Providers {
		for id, m := range p.Models {
			if m.Type == "" {
				m.Type = DetectModelType(m.ID)
			}

			if m.Type == "" {
				m.Type = DetectModelType(id)
			}

			context := modelContext{
				ID: m.ID,

				Type: m.Type,

				Name:        m.Name,
				Description: m.Description,
			}

			limit := m.Limit

			if limit == nil {
				limit = p.Limit
			}

			if limit != nil {
				context.Limiter = rate.NewLimiter(rate.Limit(*limit), *limit)
			}

			switch context.Type {
			case ModelTypeCompleter:
				completer, err := createCompleter(p, context)

				if err != nil {
					return err
				}

				cfg.RegisterCompleter(p.Type, id, completer)

			case ModelTypeEmbedder:
				embedder, err := createEmbedder(p, context)

				if err != nil {
					return err
				}

				cfg.RegisterEmbedder(p.Type, id, embedder)
				cfg.RegisterReranker(p.Type, id, provider.FromEmbedder(embedder))

			case ModelTypeReranker:
				reranker, err := createReranker(p, context)

				if err != nil {
					return err
				}

				cfg.RegisterReranker(p.Type, id, reranker)

			case ModelTypeRenderer:
				renderer, err := createRenderer(p, context)

				if err != nil {
					return err
				}

				cfg.RegisterRenderer(p.Type, id, renderer)

			case ModelTypeSynthesizer:
				synthesizer, err := createSynthesizer(p, context)

				if err != nil {
					return err
				}

				cfg.RegisterSynthesizer(p.Type, id, synthesizer)

			case ModelTypeTranscriber:
				transcriber, err := createTranscriber(p, context)

				if err != nil {
					return err
				}

				cfg.RegisterTranscriber(p.Type, id, transcriber)

			default:
				return errors.New("invalid model type: " + id)
			}
		}
	}

	return nil
}

type providerConfig struct {
	Type string `yaml:"type"`

	URL   string `yaml:"url"`
	Token string `yaml:"token"`

	Limit *int `yaml:"limit"`

	Models providerModelsConfig `yaml:"models"`
}

type providerModelsConfig map[string]modelConfig

func (c *providerModelsConfig) UnmarshalYAML(value *yaml.Node) error {
	var config map[string]modelConfig

	if err := value.Decode(&config); err == nil {
		for id, model := range config {
			if model.ID == "" {
				model.ID = id
			}

			config[id] = model
		}

		*c = config
		return nil
	}

	var list []string

	if err := value.Decode(&list); err == nil {
		config = make(map[string]modelConfig)

		for _, id := range list {
			config[id] = modelConfig{
				ID: id,
			}
		}

		*c = config
		return nil
	}

	return errors.New("invalid models config")
}
