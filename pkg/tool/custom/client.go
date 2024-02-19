package custom

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/adrianliechti/llama/pkg/jsonschema"
	"github.com/adrianliechti/llama/pkg/tool"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	_ tool.Tool = (*Client)(nil)
)

type Client struct {
	url string

	client ToolClient

	name        string
	description string
	parameters  *jsonschema.Definition
}

type Option func(*Client)

func New(url string, options ...Option) (*Client, error) {
	if url == "" || !strings.HasPrefix(url, "grpc://") {
		return nil, errors.New("invalid url")
	}

	c := &Client{
		url: url,
	}

	for _, option := range options {
		option(c)
	}

	url = strings.TrimPrefix(c.url, "grpc://")

	conn, err := grpc.Dial(url,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	if err != nil {
		return nil, err
	}

	c.client = NewToolClient(conn)

	info, err := c.client.Info(context.Background(), &InfoRequest{})

	if err != nil {
		return nil, err
	}

	c.name = info.Name
	c.description = info.Description

	if info.Schema != "" {
		var val jsonschema.Definition

		if err := json.Unmarshal([]byte(info.Schema), &val); err != nil {
			return nil, err
		}

		c.parameters = &val
	}

	return c, nil
}

func (c *Client) Name() string {
	return c.name
}

func (c *Client) Description() string {
	return c.description
}

func (c *Client) Parameters() jsonschema.Definition {
	if c.parameters == nil {
		return jsonschema.Definition{}
	}

	return *c.parameters
}

func (c *Client) Execute(ctx context.Context, parameters map[string]any) (any, error) {
	parameter, err := json.Marshal(parameters)

	if err != nil {
		return nil, err
	}

	data, err := c.client.Execute(ctx, &ExecuteRequest{
		Parameter: string(parameter),
	})

	if err != nil {
		return nil, err
	}

	var result map[string]any

	if err := json.Unmarshal([]byte(data.Content), &result); err != nil {
		return nil, err
	}

	return result, nil
}
