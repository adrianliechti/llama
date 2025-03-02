package llama

import (
	"errors"
	"strings"

	"github.com/adrianliechti/wingman/pkg/provider/jina"
)

type Reranker = jina.Reranker

func NewReranker(url, model string, options ...Option) (*Reranker, error) {
	if url == "" {
		return nil, errors.New("url is required")
	}

	url = strings.TrimRight(url, "/")
	url = strings.TrimSuffix(url, "/v1")

	cfg := &Config{}

	for _, option := range options {
		option(cfg)
	}

	opts := []jina.Option{}

	if cfg.client != nil {
		opts = append(opts, jina.WithClient(cfg.client))
	}

	return jina.NewReranker(url, model, opts...)
}
