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

	Parameters any
}

type Image struct {
	ID string

	Name    string
	Content io.ReadCloser
}

type Result struct {
	Content string
	Score   float64
}

type Usage struct {
	InputTokens  int
	OutputTokens int
}
