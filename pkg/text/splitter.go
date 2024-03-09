package text

import (
	"strings"
)

type Splitter struct {
	ChunkSize int

	Normalize  bool
	Separators []string
}

func NewSplitter() *Splitter {
	return &Splitter{
		ChunkSize: 1500,
		Normalize: true,
		Separators: []string{
			"\n\n",
			"\n",
			" ",
			"",
		},
	}
}

func (s *Splitter) Split(text string) []string {
	if s.Normalize {
		text = Normalize(text)
	}

	sep := s.textSeparator(text)
	parts := strings.Split(text, sep)

	var chunks []string
	var current []string

	for _, part := range parts {
		if len(part) < s.ChunkSize {
			current = append(current, part)
			continue
		}

		if len(current) > 0 {
			chunks = append(chunks, s.combineChunks(current, sep)...)
			clear(current)
		}

		chunks = append(chunks, s.Split(part)...)
	}

	if len(current) > 0 {
		chunks = append(chunks, s.combineChunks(current, sep)...)
	}

	var result []string

	for _, c := range chunks {
		c := strings.TrimSpace(c)

		if c == "" {
			continue
		}

		result = append(result, c)
	}

	return result
}

func (s *Splitter) combineChunks(chunks []string, sep string) []string {
	var result []string

	var chunk string

	for _, c := range chunks {
		if len(chunk)+len(c) > s.ChunkSize {
			result = append(result, chunk)
			chunk = ""
		}

		chunk += sep + c
	}

	if chunk != "" {
		result = append(result, chunk)
	}

	return result
}

func (s *Splitter) textSeparator(text string) string {
	for _, sep := range s.Separators {
		if strings.Contains(text, sep) {
			return sep
		}
	}

	return ""
}
