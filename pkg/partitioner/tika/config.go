package tika

import (
	"net/http"
)

var SupportedExtensions = []string{
	".doc", ".docx",
	".jpg", ".jpeg",
	".pdf",
	".png",
	".ppt", ".pptx",
	".xls", ".xlsx",
}

type Config struct {
	url string

	client *http.Client

	chunkSize    int
	chunkOverlap int
}

type Option func(*Config)

func WithClient(client *http.Client) Option {
	return func(c *Config) {
		c.client = client
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
