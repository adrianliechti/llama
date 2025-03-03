package api

import (
	"net/http"

	"github.com/adrianliechti/wingman/pkg/provider"
)

type Provider interface {
	http.Handler
}

type Schema = provider.Schema
