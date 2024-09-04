package azure

import (
	"net/http"
)

// https://learn.microsoft.com/en-us/azure/ai-services/document-intelligence/concept-layout?view=doc-intel-4.0.0&tabs=sample-code#input-requirements
var SupportedExtensions = []string{
	".pdf",

	".jpeg", ".jpg",
	".png",
	".bmp",
	".tiff",
	".heif",

	".docx",
	".pptx",
	".xlsx",
}

type Option func(*Client)

func WithClient(client *http.Client) Option {
	return func(c *Client) {
		c.client = client
	}
}

func WithToken(token string) Option {
	return func(c *Client) {
		c.token = token
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
