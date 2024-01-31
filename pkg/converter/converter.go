package converter

import (
	"io"
)

type ConvertOptions struct {
}

type File struct {
	ID string

	Name    string
	Content io.Reader
}

type Text struct {
}
