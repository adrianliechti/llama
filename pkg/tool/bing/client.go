package bing

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"

	"github.com/adrianliechti/llama/pkg/jsonschema"
	"github.com/adrianliechti/llama/pkg/tool"
)

var _ tool.Tool = &Tool{}

type Tool struct {
	client *http.Client
	token  string
}

func New(token string, options ...Option) (*Tool, error) {
	t := &Tool{
		client: http.DefaultClient,
		token:  token,
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
	return "bing"
}

func (t *Tool) Description() string {
	return "Search online if the requested information cannot be found in the language model or the information could be present in a time after the language model was trained."
}

func (*Tool) Parameters() jsonschema.Definition {
	return jsonschema.Definition{
		Type: jsonschema.DataTypeObject,

		Properties: map[string]jsonschema.Definition{
			"query": {
				Type:        jsonschema.DataTypeString,
				Description: "the text to search online to get the necessary information",
			},
		},

		Required: []string{"query"},
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
			Title:    p.Name,
			Content:  p.Snippet,
			Location: p.URL,
		}

		results = append(results, result)
	}

	return results, nil
}
