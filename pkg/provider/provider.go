package provider

import (
	"io"
)

type Provider = any

type Model struct {
	ID string
}

type File struct {
	ID string

	Name    string
	Content io.Reader
}

type Tool struct {
	Name        string
	Description string

	Parameters map[string]any
}

type Usage struct {
	InputTokens  int
	OutputTokens int
}
