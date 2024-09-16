package jina_test

import (
	"context"
	"testing"

	"github.com/adrianliechti/llama/pkg/provider/jina"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestEmbedder(t *testing.T) {
	ctx := context.Background()

	server, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		Started: true,

		ContainerRequest: testcontainers.ContainerRequest{
			Image: "ghcr.io/adrianliechti/llama-embeddings",

			Mounts: testcontainers.Mounts(
				testcontainers.ContainerMount{
					Target: "/app/.cache/huggingface",
					Source: testcontainers.DockerVolumeMountSource{
						Name: "huggingface",
					},
				},
			),

			ExposedPorts: []string{"8000/tcp"},

			WaitingFor: wait.ForLog("Application startup complete"),
		},
	})

	require.NoError(t, err)

	url, err := server.Endpoint(ctx, "")
	require.NoError(t, err)

	e, err := jina.NewEmbedder("http://" + url)
	require.NoError(t, err)

	result, err := e.Embed(ctx, "Hello, World!")
	require.NoError(t, err)

	require.NotEmpty(t, result.Data)
}
