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

func WithStrategy(strategy Strategy) Option {
	return func(c *Client) {
		c.strategy = strategy
	}
}
