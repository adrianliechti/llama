package text

// https://docs.unstructured.io/api-reference/api-services/supported-file-types
var SupportedExtensions = []string{
	".csv",
	".md",
	".rst",
	".tsv",
	".txt",
}

type Config struct {
	chunkSize    int
	chunkOverlap int
}

type Option func(*Config)

func WithChunkSize(size int) Option {
	return func(c *Config) {
		c.chunkSize = size
	}
}

func WithChunkOverlap(overlap int) Option {
	return func(c *Config) {
		c.chunkOverlap = overlap
	}
}
