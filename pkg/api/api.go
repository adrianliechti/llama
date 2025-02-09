package api

import (
	"net/http"

	"github.com/adrianliechti/llama/pkg/provider"
)

type Provider interface {
	http.Handler
}

type Schema = provider.Schema
