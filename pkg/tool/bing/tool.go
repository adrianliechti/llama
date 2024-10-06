package bing

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"

	"github.com/adrianliechti/llama/pkg/tool"
)

var _ tool.Tool = &Tool{}

type Tool struct {
	name        string
	description string

	token  string
	client *http.Client
}

func New(token string, options ...Option) (*Tool, error) {
	t := &Tool{
		name:        "bing",
		description: "Search online if the requested information cannot be found in the language model or the information could be present in a time after the language model was trained.",

		token:  token,
		client: http.DefaultClient,
	}

	for _, option := range options {
		option(t)
	}

	if t.token == "" {
		return nil, errors.New("invalid token")
	}

	return t, nil
}

func (t *Tool) Name() string {
	return t.name
}

func (t *Tool) Description() string {
	return t.description
}

func (*Tool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",

		"properties": map[string]any{
			"query": map[string]any{
				"type":        "string",
				"description": "the text to search online to get the necessary information",
			},
		},

		"required": []string{"query"},
	}
}

func (t *Tool) Execute(ctx context.Context, parameters map[string]any) (any, error) {
	query, ok := parameters["query"].(string)

	if !ok {
		return nil, errors.New("missing query parameter")
	}

	u, _ := url.Parse("https://api.bing.microsoft.com/v7.0/search")

	values := u.Query()
	values.Set("q", query)

	u.RawQuery = values.Encode()

	req, _ := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	req.Header.Set("Ocp-Apim-Subscription-Key", t.token)

	resp, err := t.client.Do(req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	var data SearchResponse

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	var results []Result

	for _, p := range data.WebPages.Value {
		result := Result{
			URL: p.URL,

			Title:   p.Name,
			Content: p.Snippet,
		}

		results = append(results, result)
	}

	return results, nil
}
