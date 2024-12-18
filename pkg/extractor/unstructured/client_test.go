package unstructured_test

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"testing"

	"github.com/adrianliechti/llama/pkg/extractor"
	"github.com/adrianliechti/llama/pkg/extractor/unstructured"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestExtract(t *testing.T) {
	ctx := context.Background()

	server, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		Started: true,

		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "quay.io/unstructured-io/unstructured-api:0.0.80",
			ExposedPorts: []string{"8000/tcp"},
			WaitingFor:   wait.ForLog("Application startup complete"),
		},
	})

	require.NoError(t, err)

	url, err := server.Endpoint(ctx, "")
	require.NoError(t, err)

	c, err := unstructured.New("http://"+url+"/general/v0/general",
		unstructured.WithStrategy(unstructured.StrategyFast),
	)

	require.NoError(t, err)

	resp, err := http.Get("https://helpx.adobe.com/pdf/acrobat_reference.pdf")
	require.NoError(t, err)
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	input := extractor.File{
		Name:    "acrobat_reference.pdf",
		Content: bytes.NewReader(data),
	}

	result, err := c.Extract(ctx, input, nil)
	require.NoError(t, err)

	require.NotEmpty(t, result.Content)
}
