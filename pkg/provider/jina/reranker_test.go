package jina_test

import (
	"context"
	"testing"

	"github.com/adrianliechti/wingman/pkg/provider/jina"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestReranker(t *testing.T) {
	ctx := context.Background()

	server, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		Started: true,

		ContainerRequest: testcontainers.ContainerRequest{
			Image: "ghcr.io/adrianliechti/wingman-reranker",

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

	r, err := jina.NewReranker("http://"+url, "")
	require.NoError(t, err)

	result, err := r.Rerank(ctx, "Hello, World!", []string{"World", "Sun", "Moon"}, nil)
	require.NoError(t, err)

	require.NotEmpty(t, result)
}
