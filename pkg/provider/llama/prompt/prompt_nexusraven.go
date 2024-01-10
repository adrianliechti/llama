package prompt

import (
	"errors"
	"strings"

	"github.com/adrianliechti/llama/pkg/provider"
)

type promptNexusRaven struct {
}

func (t *promptNexusRaven) Prompt(system string, messages []provider.Message, options *TemplateOptions) (string, error) {
	if options == nil {
		options = new(TemplateOptions)
	}

	message := messages[len(messages)-1]

	if message.Role != provider.MessageRoleUser {
		return "", errors.New("last message must be from user")
	}

	var prompt strings.Builder

	for _, function := range options.Functions {
		prompt.WriteString("Function:\n")
		prompt.WriteString("def " + function.Name + "(query):\n")
		prompt.WriteString("    \"\"\"\n")
		prompt.WriteString("    " + function.Description + "\n")
		prompt.WriteString("    \"\"\"\n\n")
	}

	prompt.WriteString("User Query: ")
	prompt.WriteString(strings.TrimSpace(message.Content))
	prompt.WriteString("<human_end>")
	prompt.WriteString("\n")

	println(prompt.String())

	return prompt.String(), nil
}

func (t *promptNexusRaven) Stop() []string {
	return []string{
		"<human_end>",
	}
}
