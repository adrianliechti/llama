package otel

import (
	"context"
	"net/http"

	sdkresource "go.opentelemetry.io/otel/sdk/resource"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func setupHTTP(_ context.Context, _ *sdkresource.Resource) error {
	rt := Transport(http.DefaultTransport)

	http.DefaultTransport = rt

	return nil
}

func Transport(rt http.RoundTripper) http.RoundTripper {
	if rt == nil {
		rt = http.DefaultTransport
	}

	return otelhttp.NewTransport(rt)
}
