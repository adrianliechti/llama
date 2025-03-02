package genaitoolbox

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/adrianliechti/wingman/pkg/tool"
)

var _ tool.Provider = (*Client)(nil)

type Client struct {
	url string
}

func New(url string, options ...Option) (*Client, error) {
	c := &Client{
		url: strings.TrimRight(url, "/"),
	}

	for _, option := range options {
		option(c)
	}

	return c, nil
}

func (c *Client) Tools(ctx context.Context) ([]tool.Tool, error) {
	url := c.url + "/api/toolset"

	req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)

	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("unable to list tools")
	}

	var toolset ToolSet

	if err := json.NewDecoder(resp.Body).Decode(&toolset); err != nil {
		return nil, err
	}

	var tools []tool.Tool

	for name, t := range toolset.Tools {
		props := map[string]any{}

		for _, p := range t.Parameters {
			props[p.Name] = map[string]any{
				"type":        p.Type,
				"description": p.Description,
			}
		}

		tools = append(tools, tool.Tool{
			Name:        name,
			Description: t.Description,

			Parameters: map[string]any{
				"type": "object",

				"properties": props,
			},
		})
	}

	return tools, nil
}

func (c *Client) Execute(ctx context.Context, name string, parameters map[string]any) (any, error) {
	url := c.url + "/api/tool/" + name + "/invoke"

	data, _ := json.Marshal(parameters)

	req, _ := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("unable to execute tool")
	}

	defer resp.Body.Close()

	var result struct {
		Result string `json:"result"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	var data1 []any

	if err := json.Unmarshal([]byte(result.Result), &data1); err == nil {
		return data1, nil
	}

	var data2 map[string]any

	if err := json.Unmarshal([]byte(result.Result), &data2); err == nil {
		return data2, nil
	}

	return nil, errors.New("no result found")
}

type ToolSet struct {
	Tools map[string]Tool `json:"tools"`
}

type Tool struct {
	Description string          `json:"description"`
	Parameters  []ToolParameter `json:"parameters"`
}

type ToolParameter struct {
	Type string `json:"type"`

	Name        string `json:"name"`
	Description string `json:"description"`
}
