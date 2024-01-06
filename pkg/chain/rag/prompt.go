package rag

import (
	"bytes"
	_ "embed"
	"text/template"

	"github.com/adrianliechti/llama/pkg/index"
)

var (
	//go:embed prompt.tmpl
	promptTemplateText string
	promptTemplate     = template.Must(template.New("prompt").Parse(promptTemplateText))
)

type promptData struct {
	Input   string
	Results []index.Result
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
