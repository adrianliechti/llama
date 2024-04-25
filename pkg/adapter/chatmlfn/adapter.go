package chatmlfn

import (
	"context"
	"encoding/json"
	"errors"
	"regexp"
	"strings"

	"github.com/adrianliechti/llama/pkg/adapter"
	"github.com/adrianliechti/llama/pkg/provider"
)

var _ adapter.Provider = &Adapter{}

// https://huggingface.co/datasets/Locutusque/function-calling-chatml
type Adapter struct {
	completer provider.Completer
}

func New(completer provider.Completer) (*Adapter, error) {
	a := &Adapter{
		completer: completer,
	}

	return a, nil
}

func (a *Adapter) Complete(ctx context.Context, messages []provider.Message, options *provider.CompleteOptions) (*provider.Completion, error) {
	var system string

	if len(messages) > 0 && messages[0].Role == provider.MessageRoleSystem {
		system = messages[0].Content
	}

	prompt, err := convertSystemPrompt(system, options.Functions)

	if err != nil {
		return nil, err
	}

	input := []provider.Message{
		{
			Role:    provider.MessageRoleSystem,
			Content: prompt,
		},
	}

	for _, m := range messages {
		if m.Role == provider.MessageRoleUser {
			input = append(input, m)
		}

		if m.Role == provider.MessageRoleAssistant {
			input = append(input, m)
		}

		if m.Role == provider.MessageRoleFunction {
			if m.Function == "" {
				return nil, errors.New("function is required")
			}

			prompt, err := convertToolPrompt(m.Function, m.Content)

			if err != nil {
				return nil, err
			}

			input = append(input, provider.Message{
				Role:    provider.MessageRoleFunction,
				Content: prompt,
			})
		}
	}

	completionOptions := &provider.CompleteOptions{
		Stream: options.Stream,
	}

	completion, err := a.completer.Complete(ctx, input, completionOptions)

	if err != nil {
		return nil, err
	}

	if call, err := extractToolCall(completion.Message); err == nil {
		completion = &provider.Completion{
			ID: completion.ID,

			Reason: provider.CompletionReasonFunction,

			Message: provider.Message{
				Role: provider.MessageRoleFunction,

				Function: call.Name,

				FunctionCalls: []provider.FunctionCall{
					*call,
				},
			},
		}
	}

	return completion, nil
}

func convertSystemPrompt(prompt string, functions []provider.Function) (string, error) {
	var result string

	result += "You are a helpful assistant with access to the following functions. Use them if required - "

	for _, f := range functions {
		if f.Name == "" {
			return "", errors.New("function name is required")
		}

		if f.Description == "" {
			return "", errors.New("function description is required")
		}

		if len(f.Parameters.Properties) == 0 {
			return "", errors.New("function parameters are required")
		}

		tool := Tool{
			Type: "function",

			Function: &ToolFunction{
				Name:        f.Name,
				Description: f.Description,
				Parameters:  f.Parameters,
			},
		}

		data, err := encodeJSON(tool)

		if err != nil {
			return "", err
		}

		result += data
	}

	if prompt != "" {
		result += "\n\n" + strings.TrimSpace(prompt)
	}

	return result, nil
}

func convertToolPrompt(name string, content string) (string, error) {
	callback := &ToolCallback{
		Name:    name,
		Content: []byte(content),
	}

	var result string

	result += "<tool_response>\n"

	data, _ := encodeJSON(callback)
	result += data

	result += "</tool_response>"

	return result, nil
}

func extractToolCall(message provider.Message) (*provider.FunctionCall, error) {
	re := regexp.MustCompile(`(?s)<tool_call>(.*?)</tool_call>`)
	match := re.FindStringSubmatch(message.Content)

	if len(match) == 0 {
		return nil, errors.New("no tool call found")
	}

	var result ToolCall

	if err := json.Unmarshal([]byte(match[1]), &result); err != nil {
		return nil, err
	}

	return &provider.FunctionCall{
		ID: result.Name,

		Name:      result.Name,
		Arguments: string(result.Arguments),
	}, nil
}
