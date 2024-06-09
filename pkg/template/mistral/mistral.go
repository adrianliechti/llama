package mistral

import (
	"bytes"
	"encoding/json"
	"errors"
	"regexp"
	"strings"

	"github.com/adrianliechti/llama/pkg/provider"
)

type Template struct {
}

func (t *Template) Stop() []string {
	return []string{
		"[INST]",
		"[/INST]",
	}
}

func (t *Template) Render(messages []provider.Message, options *provider.CompleteOptions) (string, error) {
	var system = ""
	var builder strings.Builder

	for _, m := range messages {
		if m.Role == provider.MessageRoleSystem {
			system = m.Content
		}

		if m.Role == provider.MessageRoleUser {
			if len(options.Functions) > 0 {
				builder.WriteString("[AVAILABLE_TOOLS] [")

				for _, f := range options.Functions {
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

					builder.WriteString(data)
				}

				builder.WriteString("][/AVAILABLE_TOOLS]")
			}

			builder.WriteString("[INST] ")

			if system != "" {
				builder.WriteString(system)
				builder.WriteString("\n")

				system = ""
			}

			builder.WriteString(m.Content)
			builder.WriteString("[/INST]")
		}

		if m.Role == provider.MessageRoleFunction {
			builder.WriteString("[TOOL_RESULTS] ")
			builder.WriteString(m.Content)
			builder.WriteString("[/TOOL_RESULTS]")
		}

		if m.Role == provider.MessageRoleSystem {
			builder.WriteString(" ")
			builder.WriteString(m.Content)
		}
	}

	return builder.String(), nil
}

func (t *Template) Parse(content string) (*provider.Message, error) {
	result := &provider.Message{
		Role: provider.MessageRoleAssistant,

		Content: content,
	}

	if calls, err := t.extractToolCalls(content); err == nil {
		result = &provider.Message{
			Role: provider.MessageRoleAssistant,

			FunctionCalls: calls,
		}
	}

	return result, nil
}

func (t *Template) extractToolCalls(s string) ([]provider.FunctionCall, error) {
	re := regexp.MustCompile(`(?s)\[TOOL_CALLS\] \[{(.*?)}\]`)
	match := re.FindStringSubmatch(s)

	if len(match) == 0 {
		return nil, errors.New("no tool call found")
	}

	content := "[{" + match[1] + "}]"
	content = strings.ReplaceAll(content, "\\n", "")
	content = strings.ReplaceAll(content, "\n", "")

	var calls []ToolCall

	if err := json.Unmarshal([]byte(content), &calls); err != nil {
		return nil, err
	}

	var results []provider.FunctionCall

	for _, c := range calls {
		result := provider.FunctionCall{
			ID: c.Name,

			Name:      c.Name,
			Arguments: string(c.Arguments),
		}

		results = append(results, result)
	}

	return results, nil
}

func encodeJSON(v any) (string, error) {
	var data bytes.Buffer

	enc := json.NewEncoder(&data)
	enc.SetEscapeHTML(false)
	enc.Encode(v)

	return data.String(), nil
}
