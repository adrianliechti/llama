package tika

import (
	"net/http"
)

var SupportedExtensions = []string{
	".pdf",

	".jpg", ".jpeg",
	".png",

	".doc", ".docx",
	".ppt", ".pptx",
	".xls", ".xlsx",
}

type Option func(*Client)

func WithClient(client *http.Client) Option {
	return func(c *Client) {
		c.client = client
	}
}

func WithChunkSize(size int) Option {
	return func(c *Client) {
		c.chunkSize = size
	}
}

func WithChunkOverlap(overlap int) Option {
	return func(c *Client) {
		c.chunkOverlap = overlap
	}
}
