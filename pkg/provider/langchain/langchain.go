package langchain

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"unicode"

	"github.com/adrianliechti/llama/pkg/provider"

	"github.com/google/uuid"
)

var (
	_ provider.Completer = (*Provider)(nil)
)

type Provider struct {
	url string

	client *http.Client
}

type Option func(*Provider)

func New(url string, options ...Option) (*Provider, error) {
	p := &Provider{
		url:    url,
		client: http.DefaultClient,
	}

	for _, option := range options {
		option(p)
	}

	if p.url == "" {
		return nil, errors.New("invalid url")
	}

	return p, nil
}

func WithClient(client *http.Client) Option {
	return func(p *Provider) {
		p.client = client
	}
}

func WithURL(url string) Option {
	return func(p *Provider) {
		p.url = url
	}
}

func (p *Provider) Complete(ctx context.Context, messages []provider.Message, options *provider.CompleteOptions) (*provider.Completion, error) {
	if options == nil {
		options = &provider.CompleteOptions{}
	}

	id := uuid.NewString()

	body, err := p.convertRunInput(messages, options)

	if err != nil {
		return nil, err
	}

	if options.Stream == nil {
		url, _ := url.JoinPath(p.url, "invoke")
		resp, err := p.client.Post(url, "application/json", jsonReader(body))

		if err != nil {
			return nil, err
		}

		if resp.StatusCode != http.StatusOK {
			return nil, errors.New("unable to complete")
		}

		defer resp.Body.Close()

		result := provider.Completion{
			ID: id,

			Reason: provider.CompletionReasonStop,

			Message: provider.Message{
				Role:    provider.MessageRoleAssistant,
				Content: "hoi",
			},
		}

		return &result, nil
	} else {
		defer close(options.Stream)

		url, _ := url.JoinPath(p.url, "stream")
		req, _ := http.NewRequestWithContext(ctx, "POST", url, jsonReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := p.client.Do(req)

		if err != nil {
			return nil, err
		}

		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, errors.New("unable to complete")
		}

		reader := bufio.NewReader(resp.Body)

		var resultText strings.Builder
		resultRole := provider.MessageRoleAssistant
		resultReason := provider.CompletionReason("")

		for i := 0; ; i++ {
			data, err := reader.ReadBytes('\n')

			if errors.Is(err, io.EOF) {
				break
			}

			if err != nil {
				return nil, err
			}

			data = bytes.TrimSpace(data)

			if bytes.EqualFold(data, []byte("event: end")) {
				resultReason = provider.CompletionReasonStop

				options.Stream <- provider.Completion{
					ID: id,

					Reason: resultReason,

					Message: provider.Message{
						Role:    resultRole,
						Content: "",
					},
				}
			}

			if bytes.HasPrefix(data, []byte("event: ")) {
				continue
			}

			// if bytes.HasPrefix(data, errorPrefix) {
			// }

			data = bytes.TrimPrefix(data, []byte("data: "))

			if len(data) == 0 {
				continue
			}

			var completion RunData

			if err := json.Unmarshal([]byte(data), &completion); err != nil {
				return nil, err
			}

			var content = completion.Content

			if i == 0 {
				content = strings.TrimLeftFunc(content, unicode.IsSpace)
			}

			resultText.WriteString(content)
			resultRole = toMessageRole(completion)

			options.Stream <- provider.Completion{
				ID: id,

				Reason: resultReason,

				Message: provider.Message{
					Role:    resultRole,
					Content: content,
				},
			}
		}

		result := provider.Completion{
			ID: id,

			Reason: resultReason,

			Message: provider.Message{
				Role:    resultRole,
				Content: resultText.String(),
			},
		}

		return &result, nil
	}
}

func (p *Provider) convertRunInput(messages []provider.Message, options *provider.CompleteOptions) (*RunInput, error) {
	if options == nil {
		options = &provider.CompleteOptions{}
	}

	result := &RunInput{}

	for _, m := range messages {
		if m.Role == provider.MessageRoleSystem {
			result.Input = append(result.Input, Input{
				Type:    InputTypeSystem,
				Content: m.Content,
			})
		}

		if m.Role == provider.MessageRoleUser {
			result.Input = append(result.Input, Input{
				Type:    InputTypeHuman,
				Content: m.Content,
			})
		}

		if m.Role == provider.MessageRoleAssistant {
			result.Input = append(result.Input, Input{
				Type:    InputTypeAI,
				Content: m.Content,
			})
		}
	}

	return result, nil
}

func toMessageRole(resp RunData) provider.MessageRole {
	if resp.Type == DataTypeAIMessageChunk {
		return provider.MessageRoleAssistant
	}

	return ""
}

func jsonReader(v any) io.Reader {
	b := new(bytes.Buffer)

	enc := json.NewEncoder(b)
	enc.SetEscapeHTML(false)

	enc.Encode(v)
	return b
}
