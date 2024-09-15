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

	c := &Config{
		client: http.DefaultClient,

		url:   url,
		token: "-",

		model: "tgi",
	}

	for _, option := range options {
		option(c)
	}

	ops := []openai.Option{}

	if c.client != nil {
		ops = append(ops, openai.WithClient(c.client))
	}

	if c.token != "" {
		ops = append(ops, openai.WithToken(c.token))
	}

	return openai.NewCompleter(url+"/v1", model, ops...)
}
