package llm

import (
	_ "embed"

	"github.com/adrianliechti/llama/pkg/classifier"
)

var (
	//go:embed prompt.tmpl
	promptTemplate string

	promptStop = []string{
		"\n###",
		"\nClass:",
		"\nCategory:",
	}
)

type promptData struct {
	Input   string
	Classes []classifier.Class
}
