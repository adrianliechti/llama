package azuretranslator

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strings"

	"github.com/adrianliechti/llama/pkg/provider"
)

type Translator struct {
	*Config
}

func NewTranslator(url string, options ...Option) (*Translator, error) {
	cfg := &Config{
		url: url,

		language: "en",

		client: http.DefaultClient,
	}

	for _, option := range options {
		option(cfg)
	}

	return &Translator{
		Config: cfg,
	}, nil
}

func (t *Translator) Translate(ctx context.Context, content string, options *provider.TranslateOptions) (*provider.Translation, error) {
	if options == nil {
		options = new(provider.TranslateOptions)
	}

	if options.Language == "" {
		options.Language = t.language
	}

	type bodyType struct {
		Text string `json:"Text"`
	}

	body := []bodyType{
		{
			Text: strings.TrimSpace(content),
		},
	}

	u, _ := url.Parse(strings.TrimRight(t.url, "/") + "/translator/text/v3.0/translate")

	query := u.Query()
	query.Set("to", options.Language)

	u.RawQuery = query.Encode()

	r, _ := http.NewRequest(http.MethodPost, u.String(), jsonReader(body))
	r.Header.Add("Ocp-Apim-Subscription-Key", t.token)
	r.Header.Add("Content-Type", "application/json")

	resp, err := t.client.Do(r)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("unable to translate")
	}

	type resultType struct {
		DetectedLanguage struct {
			Language string  `json:"language"`
			Score    float64 `json:"score"`
		} `json:"detectedLanguage"`

		Translations []struct {
			Text string `json:"text"`
			To   string `json:"to"`
		} `json:"translations"`
	}

	var result []resultType

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if len(result) == 0 || len(result[0].Translations) == 0 {
		return nil, errors.New("unable to translate content")
	}

	return &provider.Translation{
		Content: result[0].Translations[0].Text,
	}, nil
}
