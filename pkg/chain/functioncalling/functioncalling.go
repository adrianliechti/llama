package functioncalling

import (
	"context"
	"errors"
	"regexp"
	"strings"

	"github.com/adrianliechti/llama/pkg/provider"
	"github.com/google/uuid"
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

	if messages[0].Role != provider.MessageRoleUser {
		return nil, errors.New("first message must be from user")
	}

	stream := options.Stream

	options.Stop = promptStop
	options.Stream = nil

	data := promptData{
		Input: strings.TrimSpace(messages[0].Content),
	}

	for _, f := range options.Functions {
		data.Functions = append(data.Functions, promptFunction{
			Name:        f.Name,
			Description: f.Description,
		})
	}

	var history strings.Builder

	for _, m := range messages {
		if m.Role == provider.MessageRoleAssistant {
			history.WriteString(strings.TrimSpace(m.Content))
			history.WriteString("\n")
		}

		if m.Role == provider.MessageRoleFunction {
			history.WriteString("Observation: ")
			history.WriteString(strings.TrimSpace(m.Content))
			history.WriteString("\n")
		}
	}

	data.History = history.String()

	prompt := executePromptTemplate(data)

	println(prompt)

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
	context := prompt + content

	println(context)

	if result, err := extractAnswer(content); err == nil {
		c := provider.Completion{
			ID:     completion.ID,
			Reason: provider.CompletionReasonStop,

			Message: provider.Message{
				Role:    provider.MessageRoleAssistant,
				Content: result,
			},
		}

		if stream != nil {
			stream <- c
			close(stream)
		}

		return &c, nil
	}

	if fn, err := extractAction(content); err == nil {
		c := provider.Completion{
			ID:     completion.ID,
			Reason: provider.CompletionReasonFunction,

			Message: provider.Message{
				Role:    provider.MessageRoleFunction,
				Content: context,

				FunctionCalls: []provider.FunctionCall{*fn},
			},
		}

		if stream != nil {
			stream <- c
			close(stream)
		}

		return &c, nil
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
				ID: uuid.NewString(),

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
