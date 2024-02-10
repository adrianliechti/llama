package react

import (
	_ "embed"
)

var (
	//go:embed system.tmpl
	systemTemplate string

	//go:embed prompt.tmpl
	promptTemplate string

	promptStop = []string{
		"\nObservation:",
	}
)

type promptData struct {
	Input string

	Tools    []promptTool
	Messages []promptMessage
}

type promptMessage struct {
	Type    string
	Content string
}

type promptTool struct {
	Name        string
	Description string
}
