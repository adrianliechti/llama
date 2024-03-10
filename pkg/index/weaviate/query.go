package weaviate

import (
	"bytes"
	_ "embed"
	"text/template"
)

var (
	//go:embed query.tmpl
	queryTemplateText string
	queryTemplate     = template.Must(template.New("query").Parse(queryTemplateText))
)

type queryData struct {
	Class string

	Query  string
	Vector []float32

	Limit *int
	Where map[string]string
}

func executeQueryTemplate(data queryData) string {
	var buffer bytes.Buffer
	queryTemplate.Execute(&buffer, data)

	return buffer.String()
}
