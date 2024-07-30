package config

import (
	"errors"
)

func (cfg *Config) registerProviders(f *configFile) error {
	for _, p := range f.Providers {
		for id, m := range p.Models {

			switch detectModelType(id) {

			case ModelTypeCompleter:
				completer, err := createCompleter(p, m.ID)

				if err != nil {
					return err
				}

				cfg.RegisterCompleter(p.Type, id, completer)

			case ModelTypeEmbedder:
				embedder, err := createEmbedder(p, m.ID)

				if err != nil {
					return err
				}

				cfg.RegisterEmbedder(p.Type, id, embedder)

			case ModelTypeRenderer:
				renderer, err := createRenderer(p, m.ID)

				if err != nil {
					return err
				}

				cfg.RegisterRenderer(p.Type, id, renderer)

			case ModelTypeSynthesizer:
				synthesizer, err := createSynthesizer(p, m.ID)

				if err != nil {
					return err
				}

				cfg.RegisterSynthesizer(p.Type, id, synthesizer)

			case ModelTypeTranscriber:
				transcriber, err := createTranscriber(p, m.ID)

				if err != nil {
					return err
				}

				cfg.RegisterTranscriber(p.Type, id, transcriber)

			case ModelTypeTranslator:
				translator, err := createTranslator(p, m.ID)

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
