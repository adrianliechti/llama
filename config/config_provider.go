package config

import (
	"errors"

	"github.com/adrianliechti/llama/pkg/limiter"
	"github.com/adrianliechti/llama/pkg/otel"
	reranker "github.com/adrianliechti/llama/pkg/reranker/adapter"
	summarizer "github.com/adrianliechti/llama/pkg/summarizer/adapter"

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

				if _, ok := completer.(limiter.Completer); !ok {
					completer = limiter.NewCompleter(context.Limiter, completer)
				}

				if _, ok := completer.(otel.Completer); !ok {
					completer = otel.NewCompleter(p.Type, id, completer)
				}

				cfg.RegisterCompleter(id, completer)
				cfg.RegisterSummarizer(id, summarizer.FromCompleter(completer))

			case ModelTypeEmbedder:
				embedder, err := createEmbedder(p, context)

				if err != nil {
					return err
				}

				if _, ok := embedder.(limiter.Embedder); !ok {
					embedder = limiter.NewEmbedder(context.Limiter, embedder)
				}

				if _, ok := embedder.(otel.Embedder); !ok {
					embedder = otel.NewEmbedder(p.Type, id, embedder)
				}

				cfg.RegisterEmbedder(id, embedder)
				cfg.RegisterReranker(id, reranker.FromEmbedder(embedder))

			case ModelTypeRenderer:
				renderer, err := createRenderer(p, context)

				if err != nil {
					return err
				}

				if _, ok := renderer.(limiter.Renderer); !ok {
					renderer = limiter.NewRenderer(context.Limiter, renderer)
				}

				if _, ok := renderer.(otel.Renderer); !ok {
					renderer = otel.NewRenderer(p.Type, id, renderer)
				}

				cfg.RegisterRenderer(id, renderer)

			case ModelTypeSynthesizer:
				synthesizer, err := createSynthesizer(p, context)

				if err != nil {
					return err
				}

				if _, ok := synthesizer.(limiter.Synthesizer); !ok {
					synthesizer = limiter.NewSynthesizer(context.Limiter, synthesizer)
				}

				if _, ok := synthesizer.(otel.Synthesizer); !ok {
					synthesizer = otel.NewSynthesizer(p.Type, id, synthesizer)
				}

				cfg.RegisterSynthesizer(id, synthesizer)

			case ModelTypeTranscriber:
				transcriber, err := createTranscriber(p, context)

				if err != nil {
					return err
				}

				if _, ok := transcriber.(limiter.Transcriber); !ok {
					transcriber = limiter.NewTranscriber(context.Limiter, transcriber)
				}

				if _, ok := transcriber.(otel.Transcriber); !ok {
					transcriber = otel.NewTranscriber(p.Type, id, transcriber)
				}

				cfg.RegisterTranscriber(id, transcriber)

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
