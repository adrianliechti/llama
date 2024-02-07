package react

import (
	_ "embed"
)

var (
	//go:embed prompt.tmpl
	promptTemplate string

	promptStop = []string{
		"\nObservation:",
	}
)

type promptData struct {
	Input string

	Messages  []promptMessage
	Functions []promptFunction
}

type promptMessage struct {
	Type    string
	Content string
}

type promptFunction struct {
	Name        string
	Description string
}
