package config

import (
	"errors"
	"strings"

	"github.com/adrianliechti/llama/pkg/index"
	"github.com/adrianliechti/llama/pkg/index/chroma"
)

func createIndex(c indexConfig) (index.Index, error) {
	switch strings.ToLower(c.Type) {
	case "chroma":
		return chromaIndex(c)

	default:
		return nil, errors.New("invalid index type: " + c.Type)
	}
}

func chromaIndex(c indexConfig) (index.Index, error) {
	return chroma.New(c.URL, c.Name)
}
