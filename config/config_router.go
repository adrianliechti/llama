package config

import (
	"errors"
	"strings"

	"github.com/adrianliechti/llama/pkg/provider"
	"github.com/adrianliechti/llama/pkg/router/roundrobin"
)

type routerConfig struct {
	Type string `yaml:"type"`

	Models []string `yaml:"models"`
}

type routerContext struct {
	Completers []provider.Completer
}

func (cfg *Config) registerRouters(f *configFile) error {
	var configs map[string]routerConfig

	if err := f.Routers.Decode(&configs); err != nil {
		return err
	}

	for _, node := range f.Routers.Content {
		id := node.Value

		config, ok := configs[node.Value]

		if !ok {
			continue
		}

		context := routerContext{}

		for _, m := range config.Models {
			completer, err := cfg.Completer(m)

			if err != nil {
				return err
			}

			context.Completers = append(context.Completers, completer)
		}

		router, err := createRouter(config, context)

		if err != nil {
			return err
		}

		if completer, ok := router.(provider.Completer); ok {
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
