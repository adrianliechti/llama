package tavily

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
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
		name:        "tavily",
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

	u, _ := url.Parse("https://api.tavily.com/search")

	body := map[string]any{
		"api_key":      t.token,
		"query":        query,
		"search_depth": "basic",
	}

	req, _ := http.NewRequestWithContext(ctx, "POST", u.String(), jsonReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := t.client.Do(req)

	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, convertError(resp)
	}

	var data SearchResult

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

func jsonReader(v any) io.Reader {
	b := new(bytes.Buffer)

	enc := json.NewEncoder(b)
	enc.SetEscapeHTML(false)

	enc.Encode(v)
	return b
}

func convertError(resp *http.Response) error {
	return errors.New(http.StatusText(resp.StatusCode))
}
