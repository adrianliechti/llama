package prompt

import (
	"bytes"
	"text/template"
)

type Prompt struct {
	tmpl *template.Template
}

func MustNew(text string) *Prompt {
	prompt, err := New(text)

	if err != nil {
		panic(err)
	}

	return prompt
}

func New(text string) (*Prompt, error) {
	tmpl, err := template.New("prompt").Parse(text)

	if err != nil {
		return nil, err
	}

	return &Prompt{
		tmpl: tmpl,
	}, nil
}

func (t *Prompt) Execute(data any) (string, error) {
	var buffer bytes.Buffer

	if err := t.tmpl.Execute(&buffer, data); err != nil {
		return "", err
	}

	return buffer.String(), nil
}
