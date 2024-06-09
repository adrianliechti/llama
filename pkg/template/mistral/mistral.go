package mistral

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"

	"github.com/adrianliechti/llama/pkg/provider"
)

type Template struct {
}

func (t *Template) Stop() []string {
	return []string{
		"<unk>",
		"<s>",
		"</s>",
		"[INST]",
		"[/INST]",
		"[TOOL_CALLS]",
		"[AVAILABLE_TOOLS]",
		"[/AVAILABLE_TOOLS]",
		"[TOOL_RESULTS]",
		"[/TOOL_RESULTS]",
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

func encodeJSON(v any) (string, error) {
	var data bytes.Buffer

	enc := json.NewEncoder(&data)
	enc.SetEscapeHTML(false)
	enc.Encode(v)

	return data.String(), nil
}
