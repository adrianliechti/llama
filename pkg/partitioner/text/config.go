package text

// https://docs.unstructured.io/api-reference/api-services/supported-file-types
var SupportedExtensions = []string{
	".csv",
	".md",
	".rst",
	".tsv",
	".txt",
}

type Option func(*Splitter)

func WithChunkSize(size int) Option {
	return func(s *Splitter) {
		s.chunkSize = size
	}
}

func WithChunkOverlap(overlap int) Option {
	return func(s *Splitter) {
		s.chunkOverlap = overlap
	}
}
