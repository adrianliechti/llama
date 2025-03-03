package json

import (
	"github.com/adrianliechti/wingman/pkg/api"
	"github.com/adrianliechti/wingman/pkg/provider"
)

type Option func(*Handler)

func WithCompleter(p provider.Completer) Option {
	return func(c *Handler) {
		c.completer = p
	}
}

func WithInputSchema(schema api.Schema) Option {
	return func(c *Handler) {
		c.input = &schema
	}
}

func WithOutputSchema(schema api.Schema) Option {
	return func(c *Handler) {
		c.output = &schema
	}
}
