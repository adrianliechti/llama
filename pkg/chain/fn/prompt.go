package fn

import (
	"bytes"
	_ "embed"
	"text/template"
)

var (
	//go:embed prompt.tmpl
	promptTemplateText string
	promptTemplate     = template.Must(template.New("prompt").Parse(promptTemplateText))

	promptStop = []string{
		"\n###",
		"\nObservation:",
	}
)

type promptData struct {
	Input     string
	Functions []promptFunction

	Messages []promptMessage
}

type promptMessage struct {
	Type    string
	Content string
}

type promptFunction struct {
	Name        string
	Description string
}

func executePromptTemplate(data promptData) string {
	var buffer bytes.Buffer
	promptTemplate.Execute(&buffer, data)

	return buffer.String()
}
