package config

import (
	"errors"
)

func (cfg *Config) registerProviders(f *configFile) error {
	for _, p := range f.Providers {
		for id, m := range p.Models {
			var err error

			context := modelContext{
				ID: m.ID,
			}

			if len(m.Stops) > 0 {
				context.Stops = m.Stops
			}

			if m.Template != "" {
				if context.Template, err = parseTemplate(m.Template); err != nil {
					return err
				}
			}

			switch detectModelType(id) {

			case ModelTypeCompleter:
				completer, err := createCompleter(p, context)

				if err != nil {
					return err
				}

				cfg.RegisterCompleter(id, completer)

			case ModelTypeEmbedder:
				embedder, err := createEmbedder(p, context)

				if err != nil {
					return err
				}

				cfg.RegisterEmbedder(id, embedder)

			case ModelTypeRenderer:
				renderer, err := createRenderer(p, context)

				if err != nil {
					return err
				}

				cfg.RegisterRenderer(id, renderer)

			case ModelTypeSynthesizer:
				synthesizer, err := createSynthesizer(p, context)

				if err != nil {
					return err
				}

				cfg.RegisterSynthesizer(id, synthesizer)

			case ModelTypeTranscriber:
				transcriber, err := createTranscriber(p, context)

				if err != nil {
					return err
				}

				cfg.RegisterTranscriber(id, transcriber)

			case ModelTypeTranslator:
				translator, err := createTranslator(p, context)

				if err != nil {
					return err
				}

				cfg.RegisterTranslator(id, translator)

			default:
				return errors.New("invalid model type: " + id)
			}
		}
	}

	return nil
}
