package provider

import (
	"io"

	"github.com/adrianliechti/llama/pkg/jsonschema"
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

type Function struct {
	Name        string
	Description string

	Parameters jsonschema.Definition
}

type Image struct {
	ID string

	Name    string
	Content io.ReadCloser
}
