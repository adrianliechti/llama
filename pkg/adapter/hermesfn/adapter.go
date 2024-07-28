package hermesfn

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

// https://github.com/NousResearch/Hermes-Function-Calling
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
	if options == nil {
		options = new(provider.CompleteOptions)
	}

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

	completionOptions := options.Clone()
	completion, err := a.completer.Complete(ctx, input, &completionOptions)

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

	if prompt != "" {
		result += strings.TrimSpace(prompt) + "\n"
	}

	result += "You are a function calling AI model. "
	result += `You are provided with function signatures within <tools></tools> XML tags. `
	result += `You may call one or more functions to assist with the user query. `
	result += `Don't make assumptions about what values to plug into functions. `

	result += `Here are the available tools:\n`
	result += `<tools>\n`

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

	result += `</tools> `

	result += `Use the following pydantic model json schema for each tool call you will make: {"properties": {"arguments": {"title": "Arguments", "type": "object"}, "name": {"title": "Name", "type": "string"}}, "required": ["arguments", "name"], "title": "FunctionCall", "type": "object"} `

	result += `For each function call return a json object with function name and arguments within <tool_call></tool_call> XML tags as follows:\n`
	result += `<tool_call>\n`
	result += `{"arguments": <args-dict>, "name": <function-name>}`
	result += `\n</tool_call>`

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

	content := match[1]
	content = strings.ReplaceAll(content, "\\n", "")
	content = strings.ReplaceAll(content, "\n", "")

	var result ToolCall

	if err := json.Unmarshal([]byte(content), &result); err != nil {
		return nil, err
	}

	return &provider.FunctionCall{
		ID: result.Name,

		Name:      result.Name,
		Arguments: string(result.Arguments),
	}, nil
}
