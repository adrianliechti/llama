package llm

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
		"\nCategory:",
	}
)

type promptData struct {
	Input      string
	Categories []Category

	Output string
}

func executePromptTemplate(data promptData) string {
	var buffer bytes.Buffer
	promptTemplate.Execute(&buffer, data)

	return buffer.String()
}
