package template

import (
	"bytes"
	"text/template"
)

type Template struct {
	tmpl *template.Template
}

func MustTemplate(text string) *Template {
	prompt, err := NewTemplate(text)

	if err != nil {
		panic(err)
	}

	return prompt
}

func NewTemplate(text string) (*Template, error) {
	tmpl, err := template.
		New("prompt").
		Funcs(map[string]any{
			"now":        now,
			"date":       date,
			"dateInZone": dateInZone,
		}).
		Parse(text)

	if err != nil {
		return nil, err
	}

	return &Template{
		tmpl: tmpl,
	}, nil
}

func (t *Template) Execute(data any) (string, error) {
	if data == nil {
		data = map[string]any{}
	}

	var buffer bytes.Buffer

	if err := t.tmpl.Execute(&buffer, data); err != nil {
		return "", err
	}

	return buffer.String(), nil
}
