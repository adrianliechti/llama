package provider

import (
	"io"
)

type Provider = any

type Model struct {
	ID string
}

type File struct {
	Name string

	ContentType string
	Content     io.Reader
}

type Tool struct {
	Name        string
	Description string

	Strict *bool

	Parameters map[string]any
}

type Schema struct {
	Name        string
	Description string

	Strict *bool

	Schema map[string]any
}

type Usage struct {
	InputTokens  int
	OutputTokens int
}
