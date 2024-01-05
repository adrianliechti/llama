package fn

import (
	"context"
	"encoding/base64"
	"errors"
	"regexp"
	"strings"

	"github.com/adrianliechti/llama/pkg/provider"
)

var (
	_ provider.Completer = &Provider{}
)

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

	var index int

	for i, m := range messages {
		if m.Role == provider.MessageRoleUser {
			index = i
		}
	}

	messages = messages[index:]

	stream := options.Stream

	options.Stop = promptStop
	options.Stream = nil

	data := promptData{}

	for _, f := range options.Functions {
		data.Functions = append(data.Functions, promptFunction{
			Name:        f.Name,
			Description: f.Description,
		})
	}

	var history strings.Builder

	for _, m := range messages {
		if m.Role == provider.MessageRoleUser {
			history.WriteString("Question: ")
			history.WriteString(strings.TrimSpace(m.Content))
			history.WriteString("\n")

			data.Input = strings.TrimSpace(m.Content)
		}

		if m.Role == provider.MessageRoleAssistant {
			if m.Content == "" {
				continue
			}

			history.WriteString("Thought: I now know the final answer.")
			history.WriteString("\n")
			history.WriteString("Final Answer: ")
			history.WriteString(strings.TrimSpace(m.Content))
			history.WriteString("\n")
		}

		if m.Role == provider.MessageRoleFunction {
			if data, err := base64.RawStdEncoding.DecodeString(m.Function); err == nil {
				history.WriteString(strings.TrimSpace(string(data)))
				history.WriteString("\n")
			}

			history.WriteString("Observation: ")
			history.WriteString(strings.TrimSpace(m.Content))
			history.WriteString("\n")
		}
	}

	data.History = history.String()

	prompt := executePromptTemplate(data)

	inputMesssages := []provider.Message{
		{
			Role:    provider.MessageRoleUser,
			Content: prompt,
		},
	}

	completion, err := p.completer.Complete(ctx, inputMesssages, options)

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

		if stream != nil {
			stream <- result
			close(stream)
		}

		return &result, nil
	}

	if fn, err := extractAction(content); err == nil {
		fn.ID = base64.RawStdEncoding.EncodeToString([]byte(content))

		result := provider.Completion{
			ID:     completion.ID,
			Reason: provider.CompletionReasonFunction,

			Message: provider.Message{
				Role: provider.MessageRoleAssistant,

				FunctionCalls: []provider.FunctionCall{*fn},
			},
		}

		if stream != nil {
			stream <- result
			close(stream)
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
