package prompt

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/adrianliechti/llama/pkg/jsonschema"
	"github.com/adrianliechti/llama/pkg/provider"
)

type promptGorilla struct {
}

func (t *promptGorilla) Prompt(system string, messages []provider.Message, options *TemplateOptions) (string, error) {
	if options == nil {
		options = new(TemplateOptions)
	}

	message := messages[len(messages)-1]

	if message.Role != provider.MessageRoleUser {
		return "", errors.New("last message must be from user")
	}

	var prompt strings.Builder

	for _, message := range messages {
		if message.Role == provider.MessageRoleUser {
			prompt.WriteString("USER: <<question>> ")
			prompt.WriteString(strings.TrimSpace(message.Content))

			if len(options.Functions) > 0 {
				prompt.WriteString("<<function>> ")

				var functions []jsonschema.FunctionDefinition

				for _, f := range options.Functions {
					function := jsonschema.FunctionDefinition{
						Name:        f.Name,
						Description: f.Description,

						Parameters: f.Parameters,
					}

					functions = append(functions, function)
				}

				data, _ := json.Marshal(functions)
				prompt.WriteString(string(data))
			}

			prompt.WriteString("\n")
		}

		if message.Role == provider.MessageRoleAssistant {
			prompt.WriteString("ASSSISTANT: ")
			prompt.WriteString(strings.TrimSpace(message.Content))
			prompt.WriteString("\n")

		}
	}

	prompt.WriteString("ASSSISTANT: ")

	println(prompt.String())

	return prompt.String(), nil
}

func (t *promptGorilla) Stop() []string {
	return []string{
		"USER:",
		"ASSSISTANT:",
	}
}
