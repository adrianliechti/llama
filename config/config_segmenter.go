package config

import (
	"errors"
	"strings"

	"github.com/adrianliechti/llama/pkg/limiter"
	"github.com/adrianliechti/llama/pkg/otel"
	"github.com/adrianliechti/llama/pkg/segmenter"
	"github.com/adrianliechti/llama/pkg/segmenter/jina"
	"github.com/adrianliechti/llama/pkg/segmenter/text"
	"github.com/adrianliechti/llama/pkg/segmenter/unstructured"
	"golang.org/x/time/rate"
)

func (cfg *Config) RegisterSegmenter(id string, p segmenter.Provider) {
	if cfg.segmenter == nil {
		cfg.segmenter = make(map[string]segmenter.Provider)
	}

	cfg.segmenter[id] = p
}

func (cfg *Config) Segmenter(id string) (segmenter.Provider, error) {
	if cfg.segmenter != nil {
		if p, ok := cfg.segmenter[id]; ok {
			return p, nil
		}
	}

	if id == "" {
		// Take any Segmenter
		for _, p := range cfg.segmenter {
			return p, nil
		}

		return text.New()
	}

	return nil, errors.New("segmenter not found: " + id)
}

type segmenterConfig struct {
	Type string `yaml:"type"`

	URL   string `yaml:"url"`
	Token string `yaml:"token"`

	Limit *int `yaml:"limit"`
}

type segmenterContext struct {
	Limiter *rate.Limiter
}

func (cfg *Config) RegisterSegmenters(f *configFile) error {
	for id, e := range f.Segmenters {
		context := segmenterContext{}

		limit := e.Limit

		if limit == nil {
			limit = e.Limit
		}

		if limit != nil {
			context.Limiter = rate.NewLimiter(rate.Limit(*limit), *limit)
		}

		segmenter, err := createSegmenter(e, context)

		if err != nil {
			return err
		}

		if _, ok := segmenter.(limiter.Segmenter); !ok {
			segmenter = limiter.NewSegmenter(context.Limiter, segmenter)
		}

		if _, ok := segmenter.(otel.Segmenter); !ok {
			segmenter = otel.NewSegmenter(id, segmenter)
		}

		cfg.RegisterSegmenter(id, segmenter)
	}

	return nil
}

func createSegmenter(cfg segmenterConfig, context segmenterContext) (segmenter.Provider, error) {
	switch strings.ToLower(cfg.Type) {

	case "jina":
		return jinaSegmenter(cfg)

	case "text":
		return textSegmenter(cfg)

	case "unstructured":
		return unstructuredSegmenter(cfg)

	default:
		return nil, errors.New("invalid segmenter type: " + cfg.Type)
	}
}

func jinaSegmenter(cfg segmenterConfig) (segmenter.Provider, error) {
	var options []jina.Option

	if cfg.Token != "" {
		options = append(options, jina.WithToken(cfg.Token))
	}

	return jina.New(cfg.URL, options...)
}

func textSegmenter(cfg segmenterConfig) (segmenter.Provider, error) {
	return text.New()
}

func unstructuredSegmenter(cfg segmenterConfig) (segmenter.Provider, error) {
	var options []unstructured.Option

	if cfg.Token != "" {
		options = append(options, unstructured.WithToken(cfg.Token))
	}

	return unstructured.New(cfg.URL, options...)
}
