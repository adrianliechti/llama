package prompt

import (
	"encoding/json"
	"errors"
	"strings"

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

	prompt.WriteString("USER: <<question>> ")
	prompt.WriteString(strings.TrimSpace(message.Content))

	if len(options.Functions) > 0 {
		prompt.WriteString(" <<function>> ")

		var functions []function

		for _, f := range options.Functions {
			function := function{
				Call: f.Name,

				Name:        f.Name,
				Description: f.Description,

				Parameters: []functionParameter{},
			}

			for k, v := range f.Parameters.Properties {
				parameter := functionParameter{
					Name: k,

					Enum:        v.Enum,
					Description: v.Description,
				}

				function.Parameters = append(function.Parameters, parameter)

			}

			functions = append(functions, function)
		}

		data, _ := json.Marshal(functions)
		prompt.WriteString(string(data))
	}

	prompt.WriteString("\n")

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

type function struct {
	Call string `json:"api_call,omitempty"`

	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`

	Parameters []functionParameter `json:"parameters,omitempty"`
}

type functionParameter struct {
	Name string `json:"name,omitempty"`

	Enum        []string `json:"enum,omitempty"`
	Description string   `json:"description,omitempty"`
}
