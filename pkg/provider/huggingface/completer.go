package huggingface

import (
	"net/http"
	"strings"

	"github.com/adrianliechti/llama/pkg/provider/openai"
)

type Completer = openai.Completer

func NewCompleter(url, model string, options ...Option) (*Completer, error) {
	if url == "" {
		url = "https://api-inference.huggingface.co/models/" + model
	}

	url = strings.TrimRight(url, "/")
	url = strings.TrimRight(url, "/v1")

	cfg := &Config{
		client: http.DefaultClient,

		url:   url,
		token: "-",

		model: "tgi",
	}

	for _, option := range options {
		option(cfg)
	}

	ops := []openai.Option{}

	if cfg.client != nil {
		ops = append(ops, openai.WithClient(cfg.client))
	}

	if cfg.token != "" {
		ops = append(ops, openai.WithToken(cfg.token))
	}

	return openai.NewCompleter(url+"/v1", model, ops...)
}
