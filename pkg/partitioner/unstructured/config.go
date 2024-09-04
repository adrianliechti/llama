package unstructured

import (
	"net/http"
)

// https://docs.unstructured.io/api-reference/api-services/supported-file-types
var SupportedExtensions = []string{
	".bmp",
	".csv",
	".doc",
	".docx",
	".eml",
	".epub",
	".heic",
	".html",
	".jpeg",
	".png",
	".md",
	".msg",
	".odt",
	".org",
	".p7s",
	".pdf",
	".png",
	".ppt",
	".pptx",
	".rst",
	".rtf",
	".tiff",
	".txt",
	".tsv",
	".xls",
	".xlsx",
	".xml",
}

type Config struct {
	client *http.Client

	url   string
	token string

	chunkSize     int
	chunkOverlap  int
	chunkStrategy string
}

type Option func(*Config)

func WithClient(client *http.Client) Option {
	return func(c *Config) {
		c.client = client
	}
}

func WithURL(url string) Option {
	return func(c *Config) {
		c.url = url
	}
}

func WithToken(token string) Option {
	return func(c *Config) {
		c.token = token
	}
}

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
