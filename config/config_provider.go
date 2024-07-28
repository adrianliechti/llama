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

				if m.Adapter != "" {
					adapter, err := createCompleterAdapter(m.Adapter, completer)

					if err != nil {
						return err
					}

					completer = adapter
				}

				cfg.RegisterCompleter(id, completer)

			case ModelTypeEmbedder:
				embedder, err := createEmbedder(p, m.ID)

				if err != nil {
					return err
				}

				cfg.RegisterEmbedder(id, embedder)

			case ModelTypeRenderer:
				renderer, err := createRenderer(p, m.ID)

				if err != nil {
					return err
				}

				cfg.RegisterRenderer(id, renderer)

			case ModelTypeSynthesizer:
				synthesizer, err := createSynthesizer(p, m.ID)

				if err != nil {
					return err
				}

				cfg.RegisterSynthesizer(id, synthesizer)

			case ModelTypeTranscriber:
				transcriber, err := createTranscriber(p, m.ID)

				if err != nil {
					return err
				}

				cfg.RegisterTranscriber(id, transcriber)

			case ModelTypeTranslator:
				translator, err := createTranslator(p, m.ID)

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
