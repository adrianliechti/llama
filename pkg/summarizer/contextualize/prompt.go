package contextualize

import (
	_ "embed"
)

var (
	//go:embed prompt.tmpl
	promptTemplate string
)

type promptData struct {
	Input    string
	Messages []promptMessage
}

type promptMessage struct {
	Role    string
	Content string
}
