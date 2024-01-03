package functioncalling

import (
	"context"
	"encoding/json"
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
		"\nObservation:",
	}

	// if messages[0].Role == provider.MessageRoleSystem {
	// 	messages = messages[1:]
	// }

	// system := provider.Message{
	// 	Role:    provider.MessageRoleSystem,
	// 	Content: promptSystem,
	// }

	// lala = append(lala, system)

	var content strings.Builder

	content.WriteString(promptSystem)
	content.WriteString("\n")
	content.WriteString("\n")

	content.WriteString("### Input:\n")
	content.WriteString(strings.TrimSpace(messages[0].Content))
	content.WriteString("\n\n")

	content.WriteString("### Output:\n")

	for _, m := range messages {
		if m.Role == provider.MessageRoleAssistant {
			text := strings.TrimSpace(m.Content)
			content.WriteString(text)
			content.WriteString("\n")
		}

		if m.Role == provider.MessageRoleFunction {
			text := strings.TrimSpace(m.Content)
			content.WriteString("Observation: ")
			content.WriteString(text)
			content.WriteString("\n")
		}
	}

	println(content.String())

	messsage := provider.Message{
		Role:    provider.MessageRoleUser,
		Content: content.String(),
	}

	// messages = append([]provider.Message{system}, messages...)

	var msgs []provider.Message
	msgs = append(msgs, messsage)

	completion, err := p.completer.Complete(ctx, msgs, options)

	if err != nil {
		return nil, err
	}

	out := completion.Message

	if fn, err := extractAction(out.Content); err == nil {
		out.FunctionCalls = append(out.FunctionCalls, *fn)
		completion.Reason = provider.CompletionReasonFunction
	}

	completion.Message = out

	dumpMessages([]provider.Message{out})

	return completion, err
}

func dumpMessages(msgs []provider.Message) {
	for _, m := range msgs {
		data, _ := json.MarshalIndent(m, "", "  ")
		println(string(data))
	}
}

func extractAction(s string) (*provider.FunctionCall, error) {
	re := regexp.MustCompile(`Action: (.*)\s+Action Input: (.*)$`)
	matches := re.FindStringSubmatch(s)

	if len(matches) != 3 {
		return nil, errors.New("invalid action")
	}

	return &provider.FunctionCall{
		ID: uuid.NewString(),

		Name:      matches[1],
		Arguments: matches[2],
	}, nil
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
