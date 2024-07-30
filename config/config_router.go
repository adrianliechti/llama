package config

import (
	"errors"
	"strings"

	"github.com/adrianliechti/llama/pkg/provider"
	"github.com/adrianliechti/llama/pkg/router/roundrobin"
)

type routerContext struct {
	Completers []provider.Completer
}

func (cfg *Config) registerRouters(f *configFile) error {
	for id, r := range f.Routers {
		context := routerContext{}

		for _, m := range r.Models {
			completer, err := cfg.Completer(m)

			if err != nil {
				return err
			}

			context.Completers = append(context.Completers, completer)
		}

		r, err := createRouter(r, context)

		if err != nil {
			return err
		}

		if completer, ok := r.(provider.Completer); ok {
			cfg.RegisterCompleter(id, completer)
		}
	}

	return nil
}

func createRouter(cfg routerConfig, context routerContext) (any, error) {
	switch strings.ToLower(cfg.Type) {
	case "roundrobin":
		return roundrobinRouter(cfg, context)

	default:
		return nil, errors.New("invalid router type: " + cfg.Type)
	}
}

func roundrobinRouter(cfg routerConfig, context routerContext) (any, error) {
	return roundrobin.NewCompleter(context.Completers...)
}
