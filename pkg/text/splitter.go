package text

import (
	"strings"
	"unicode/utf8"
)

type Splitter struct {
	ChunkSize    int
	ChunkOverlap int

	Separators []string

	LenFunc func(string) int
}

func NewSplitter() Splitter {
	s := Splitter{
		ChunkSize:    1500,
		ChunkOverlap: 0,

		Separators: []string{
			"\n\n",
			"\n",
			" ",
			"",
		},

		LenFunc: utf8.RuneCountInString,
	}

	return s
}

func (s *Splitter) Split(text string) []string {
	text = Normalize(text)
	separator := s.textSeparator(text)

	result := make([]string, 0)
	chunks := make([]string, 0)

	for _, split := range strings.Split(text, separator) {
		if s.LenFunc(split) < s.ChunkSize {
			chunks = append(chunks, split)
			continue
		}

		if len(chunks) > 0 {
			result = append(result, s.mergeSplits(chunks, separator)...)
			clear(chunks)
		}

		result = append(result, s.Split(split)...)
	}

	if len(chunks) > 0 {
		result = append(result, s.mergeSplits(chunks, separator)...)
	}

	return result
}

func (s *Splitter) mergeSplits(splits []string, separator string) []string {
	total := 0

	result := make([]string, 0)
	current := make([]string, 0)

	for _, split := range splits {
		splitlength := total + s.LenFunc(split)

		if len(current) != 0 {
			splitlength += s.LenFunc(separator)
		}

		if splitlength > s.ChunkSize && len(current) > 0 {
			doc := joinDocs(current, separator)

			if doc != "" {
				result = append(result, doc)
			}

			for shouldPop(s.ChunkOverlap, s.ChunkSize, total, s.LenFunc(split), s.LenFunc(separator), len(current)) {
				total -= s.LenFunc(current[0])

				if len(current) > 1 {
					total -= s.LenFunc(separator)
				}

				current = current[1:]
			}
		}

		current = append(current, split)
		total += s.LenFunc(split)

		if len(current) > 1 {
			total += s.LenFunc(separator)
		}
	}

	doc := joinDocs(current, separator)

	if doc != "" {
		result = append(result, doc)
	}

	return result
}

func joinDocs(docs []string, separator string) string {
	return strings.TrimSpace(strings.Join(docs, separator))
}

func shouldPop(chunkOverlap, chunkSize, total, splitLen, separatorLen, currentDocLen int) bool {
	docsNeededToAddSep := 2

	if currentDocLen < docsNeededToAddSep {
		separatorLen = 0
	}

	return currentDocLen > 0 && (total > chunkOverlap || (total+splitLen+separatorLen > chunkSize && total > 0))
}

func (s *Splitter) textSeparator(text string) string {
	for _, sep := range s.Separators {
		if strings.Contains(text, sep) {
			return sep
		}
	}

	if len(s.Separators) > 0 {
		return s.Separators[len(s.Separators)-1]
	}

	return ""
}
