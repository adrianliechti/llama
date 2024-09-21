package searxng

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
	url string

	client *http.Client
}

func New(url string, options ...Option) (*Tool, error) {
	t := &Tool{
		client: http.DefaultClient,

		url: url,
	}

	for _, option := range options {
		option(t)
	}

	return t, nil
}

func (t *Tool) Name() string {
	return "searxng"
}

func (t *Tool) Description() string {
	return "Search online if the requested information cannot be found in the language model or the information could be present in a time after the language model was trained."
}

func (*Tool) Parameters() any {
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

	url, _ := url.Parse(t.url)
	url = url.JoinPath("/search")

	values := url.Query()
	values.Set("q", query)
	values.Set("format", "json")
	values.Set("safesearch", "0")

	url.RawQuery = values.Encode()

	req, _ := http.NewRequestWithContext(ctx, "GET", url.String(), nil)

	resp, err := t.client.Do(req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, errors.New("failed to fetch search results")
	}

	var data SearchResponse

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	var results []Result

	for _, r := range data.Results {
		result := Result{
			URL: r.URL,

			Title:   r.Title,
			Content: r.Content,
		}

		results = append(results, result)
	}

	return results, nil
}
