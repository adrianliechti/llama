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

	if len(options.Functions) > 0 {
		println("change prompt: inject functions")
	}

	options.Stop = []string{
		"\n###",
		"\nObservation:",
	}

	var input strings.Builder

	input.WriteString(promptSystem)
	input.WriteString("\n")
	input.WriteString("\n")

	input.WriteString("### Input:\n")
	input.WriteString(strings.TrimSpace(messages[0].Content))
	input.WriteString("\n\n")

	input.WriteString("### Output:\n")

	for _, m := range messages {
		if m.Role == provider.MessageRoleAssistant {
			input.WriteString(strings.TrimSpace(m.Content))
			input.WriteString("\n")
		}

		if m.Role == provider.MessageRoleFunction {
			input.WriteString("Observation: ")
			input.WriteString(strings.TrimSpace(m.Content))
			input.WriteString("\n")
		}
	}

	println(input.String())

	inputMesssages := []provider.Message{
		{
			Role:    provider.MessageRoleUser,
			Content: input.String(),
		},
	}

	completion, err := p.completer.Complete(ctx, inputMesssages, options)

	if err != nil {
		return nil, err
	}

	content := strings.TrimSpace(completion.Message.Content)
	context := input.String() + content

	println(context)

	if result, err := extractAnswer(content); err == nil {
		return &provider.Completion{
			ID:     completion.ID,
			Reason: provider.CompletionReasonStop,

			Message: provider.Message{
				Role:    provider.MessageRoleAssistant,
				Content: result,
			},
		}, nil
	}

	if fn, err := extractAction(content); err == nil {
		return &provider.Completion{
			ID:     completion.ID,
			Reason: provider.CompletionReasonFunction,

			Message: provider.Message{
				Role:    provider.MessageRoleFunction,
				Content: context,

				FunctionCalls: []provider.FunctionCall{*fn},
			},
		}, nil
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
			return match[0], nil
		}
	}

	return "", errors.New("no answer found")
}

var promptSystem = `Below is an instruction that describes a task, paired with an input that provides further context. Write a response that appropriately completes the request.

### Instruction:
Answer the following questions as best you can. You have access to the following tools:

Google Search: A wrapper around Google Search. Useful for when you need to answer questions about current events. The input is the question to search relavant information.

Strictly use the following format:

Question: the input question you must answer
Thought: you should always think about what to do
Action: the action to take, should be one of [Google Search]
Action Input: the input to the action, should be a question.
Observation: the result of the action
... (this Thought/Action/Action Input/Observation can repeat N times)
Thought: I now know the final answer
Final Answer: the final answer to the original input question

For examples:
Question: How old is CEO of Microsoft wife?
Thought: First, I need to find who is the CEO of Microsoft.
Action: Google Search
Action Input: Who is the CEO of Microsoft?
Observation: Satya Nadella is the CEO of Microsoft.
Thought: Now, I should find out Satya Nadella's wife.
Action: Google Search
Action Input: Who is Satya Nadella's wife?
Observation: Satya Nadella's wife's name is Anupama Nadella.
Thought: Then, I need to check Anupama Nadella's age.
Action: Google Search
Action Input: How old is Anupama Nadella?
Observation: Anupama Nadella's age is 50.
Thought: I now know the final answer.
Final Answer: Anupama Nadella is 50 years old.`
