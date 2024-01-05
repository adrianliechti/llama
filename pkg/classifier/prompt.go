package classifier

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
		"\nClass:",
	}
)

type promptData struct {
	Input   string
	Classes []Class

	Output string
}

func executePromptTemplate(data promptData) string {
	var buffer bytes.Buffer
	promptTemplate.Execute(&buffer, data)

	return buffer.String()
}
