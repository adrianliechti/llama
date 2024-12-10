package replicate

import (
	"context"

	"github.com/replicate/replicate-go"
)

type Client struct {
	*Config
	client *replicate.Client
}

type PredictionInput = replicate.PredictionInput
type PredictionOutput = replicate.PredictionOutput

type FileOutput = replicate.FileOutput

func New(model string, options ...Option) (*Client, error) {
	cfg := &Config{
		model: model,
	}

	for _, option := range options {
		option(cfg)
	}

	client, err := replicate.NewClient(cfg.Options()...)

	if err != nil {
		return nil, err
	}

	return &Client{
		Config: cfg,
		client: client,
	}, nil
}

func (c *Client) Run(ctx context.Context, input PredictionInput) (PredictionOutput, error) {
	return c.client.RunWithOptions(ctx, c.model, input, nil, replicate.WithBlockUntilDone(), replicate.WithFileOutput())
}
