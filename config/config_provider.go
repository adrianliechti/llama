package config

import (
	"errors"
)

func (cfg *Config) registerProviders(f *configFile) error {
	for _, p := range f.Providers {
		for id, m := range p.Models {
			context := modelContext{
				ID: m.ID,

				Type: detectModelType(id),

				Name:        m.Name,
				Description: m.Description,
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

			case ModelTypeTranslator:
				translator, err := createTranslator(p, context)

				if err != nil {
					return err
				}

				cfg.RegisterTranslator(p.Type, id, translator)

			default:
				return errors.New("invalid model type: " + id)
			}
		}
	}

	return nil
}
