package deepl

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strings"

	"github.com/adrianliechti/llama/pkg/provider"
)

func (c *Client) Translate(ctx context.Context, content string, options *provider.TranslateOptions) (*provider.Translation, error) {
	if options == nil {
		options = new(provider.TranslateOptions)
	}

	if options.Language == "" {
		options.Language = c.language
	}

	type bodyType struct {
		Text       []string `json:"text"`
		TargetLang string   `json:"target_lang"`
	}

	body := bodyType{
		Text: []string{
			strings.TrimSpace(content),
		},

		TargetLang: options.Language,
	}

	u, _ := url.JoinPath(c.url, "/v2/translate")
	r, _ := http.NewRequest(http.MethodPost, u, jsonReader(body))
	r.Header.Add("Authorization", "DeepL-Auth-Key "+c.token)
	r.Header.Add("Content-Type", "application/json")

	resp, err := c.client.Do(r)

	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("unable to translate")
	}

	defer resp.Body.Close()

	type resultType struct {
		Translations []struct {
			DetectedSourceLanguage string `json:"detected_source_language"`
			Text                   string `json:"text"`
		} `json:"translations"`
	}

	var result resultType

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if len(result.Translations) == 0 {
		return nil, errors.New("unable to translate content")
	}

	return &provider.Translation{
		Content: result.Translations[0].Text,
	}, nil
}
