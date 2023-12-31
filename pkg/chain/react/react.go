package react

import (
	"context"
	"encoding/base64"
	"errors"
	"regexp"
	"strings"

	"github.com/adrianliechti/llama/pkg/provider"
)

var _ provider.Completer = &Provider{}

type Provider struct {
	completer provider.Completer
}

type Option func(*Provider)

func New(options ...Option) (*Provider, error) {
	p := &Provider{}

	for _, option := range options {
		option(p)
	}

	if p.completer == nil {
		return nil, errors.New("missing completer provider")
	}

	return p, nil
}

func WithCompleter(completer provider.Completer) Option {
	return func(p *Provider) {
		p.completer = completer
	}
}

func (p *Provider) Complete(ctx context.Context, messages []provider.Message, options *provider.CompleteOptions) (*provider.Completion, error) {
	if options == nil {
		options = new(provider.CompleteOptions)
	}

	if len(options.Functions) == 0 {
		return p.completer.Complete(ctx, messages, options)
	}

	data := promptData{}

	for _, f := range options.Functions {
		data.Functions = append(data.Functions, promptFunction{
			Name:        f.Name,
			Description: f.Description,
		})
	}

	var history []promptMessage

	for _, m := range messages {
		if m.Role == provider.MessageRoleUser {
			history = append(history, promptMessage{
				Type:    "Question",
				Content: strings.TrimSpace(m.Content),
			})

			data.Input = strings.TrimSpace(m.Content)
		}

		if m.Role == provider.MessageRoleAssistant {
			if m.Content == "" {
				continue
			}

			history = append(history,
				promptMessage{
					Type:    "Thought",
					Content: "I now know the final answer.",
				},
				promptMessage{
					Type:    "Final Answer",
					Content: strings.TrimSpace(m.Content),
				})
		}

		if m.Role == provider.MessageRoleFunction {
			if data, err := base64.RawStdEncoding.DecodeString(m.Function); err == nil {
				parts := strings.Split(strings.TrimSpace(string(data)), "\n")

				for _, part := range parts {
					part = strings.TrimSpace(part)

					if part == "" {
						continue
					}

					parts := strings.SplitN(part, ":", 2)

					if len(parts) != 2 {
						continue
					}

					history = append(history,
						promptMessage{
							Type:    strings.TrimSpace(parts[0]),
							Content: strings.TrimSpace(parts[1]),
						},
					)
				}
			}

			history = append(history,
				promptMessage{
					Type:    "Observation",
					Content: strings.TrimSpace(m.Content),
				},
			)
		}
	}

	data.Messages = history

	prompt, err := promptTemplate.Execute(data)

	if err != nil {
		return nil, err
	}

	println(prompt)

	inputMesssages := []provider.Message{
		{
			Role:    provider.MessageRoleUser,
			Content: prompt,
		},
	}

	completion, err := p.completer.Complete(ctx, inputMesssages, nil)

	if err != nil {
		return nil, err
	}

	content := strings.TrimSpace(completion.Message.Content)

	if answer, err := extractAnswer(content); err == nil {
		result := provider.Completion{
			ID:     completion.ID,
			Reason: provider.CompletionReasonStop,

			Message: provider.Message{
				Role:    provider.MessageRoleAssistant,
				Content: answer,
			},
		}

		return &result, nil
	}

	if action, err := extractAction(content); err == nil {
		action.ID = base64.RawStdEncoding.EncodeToString([]byte(content))

		result := provider.Completion{
			ID:     completion.ID,
			Reason: provider.CompletionReasonFunction,

			Message: provider.Message{
				Role: provider.MessageRoleAssistant,

				FunctionCalls: []provider.FunctionCall{*action},
			},
		}

		return &result, nil
	}

	return nil, errors.New("no answer found")
}

func extractAction(s string) (*provider.FunctionCall, error) {
	re := regexp.MustCompile(`Action: (.*)\s+Action Input: (.*)`)
	matches := re.FindAllStringSubmatch(s, -1)

	if len(matches) > 0 {
		match := matches[len(matches)-1]

		if len(match) == 3 {
			args := "{\"query\": \"" + match[2] + "\"}"

			return &provider.FunctionCall{
				Name:      match[1],
				Arguments: args,
			}, nil
		}
	}

	return nil, errors.New("no action found")
}

func extractAnswer(s string) (string, error) {
	re := regexp.MustCompile(`Final Answer: (.*)`)
	matches := re.FindAllStringSubmatch(s, -1)

	if len(matches) > 0 {
		match := matches[len(matches)-1]

		if len(match) == 2 {
			return match[1], nil
		}
	}

	return "", errors.New("no answer found")
}
