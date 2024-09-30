package duckduckgo

import (
	"bufio"
	"context"
	"errors"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/adrianliechti/llama/pkg/text"
	"github.com/adrianliechti/llama/pkg/tool"
)

var _ tool.Tool = &Tool{}

type Tool struct {
	name        string
	description string

	client *http.Client
}

func New(options ...Option) (*Tool, error) {
	t := &Tool{
		name:        "duckduckgo",
		description: "Search online if the requested information cannot be found in the language model or the information could be present in a time after the language model was trained.",

		client: http.DefaultClient,
	}

	for _, option := range options {
		option(t)
	}

	return t, nil
}

func (t *Tool) Name() string {
	return t.name
}

func (t *Tool) Description() string {
	return t.description
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

	url, _ := url.Parse("https://duckduckgo.com/html/")

	values := url.Query()
	values.Set("q", query)

	url.RawQuery = values.Encode()

	req, _ := http.NewRequestWithContext(ctx, "GET", url.String(), nil)
	req.Header.Set("Referer", "https://www.duckduckgo.com/")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Sec-Fetch-Dest", "document")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	req.Header.Set("Sec-Fetch-Site", "cross-site")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.3.1 Safari/605.1.15")

	resp, err := t.client.Do(req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	var results []Result

	regexLink := regexp.MustCompile(`href="([^"]+)"`)
	regexSnippet := regexp.MustCompile(`<[^>]*>`)

	scanner := bufio.NewScanner(resp.Body)

	var resultURL string
	var resultTitle string
	var resultSnippet string

	for scanner.Scan() {
		line := scanner.Text()

		if strings.Contains(line, "result__a") {
			snippet := regexSnippet.ReplaceAllString(line, "")
			snippet = text.Normalize(snippet)

			resultTitle = snippet
		}

		if strings.Contains(line, "result__url") {
			links := regexLink.FindStringSubmatch(line)

			if len(links) >= 2 {
				resultURL = links[1]
			}
		}

		if strings.Contains(line, "result__snippet") {
			snippet := regexSnippet.ReplaceAllString(line, "")
			snippet = text.Normalize(snippet)

			resultSnippet = snippet
		}

		if resultSnippet == "" {
			continue
		}

		result := Result{
			URL:     resultURL,
			Title:   resultTitle,
			Content: resultSnippet,
		}

		results = append(results, result)

		resultURL = ""
		resultTitle = ""
		resultSnippet = ""
	}

	return results, nil
}
