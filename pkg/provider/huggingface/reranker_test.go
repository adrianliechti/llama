package huggingface_test

import (
	"context"
	"testing"

	"github.com/adrianliechti/llama/pkg/provider/huggingface"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestReranker(t *testing.T) {
	ctx := context.Background()

	server, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		Started: true,

		ContainerRequest: testcontainers.ContainerRequest{
			Image:         "ghcr.io/huggingface/text-embeddings-inference:cpu-1.5",
			ImagePlatform: "linux/amd64",

			Cmd: []string{"--model-id", "BAAI/bge-reranker-base"},

			Mounts: testcontainers.Mounts(
				testcontainers.ContainerMount{
					Target: "/data",
					Source: testcontainers.DockerVolumeMountSource{
						Name: "huggingface",
					},
				},
			),

			ExposedPorts: []string{"80/tcp"},

			WaitingFor: wait.ForLog("Ready"),
		},
	})

	require.NoError(t, err)

	url, err := server.Endpoint(ctx, "")
	require.NoError(t, err)

	e, err := huggingface.NewReranker("http://"+url, "")
	require.NoError(t, err)

	result, err := e.Rerank(ctx, "What is Deep Learning", []string{"Deep learning is a type of machine learning that uses artificial neural networks to learn from data.", "Deep Learning is Not All You Need"}, nil)
	require.NoError(t, err)

	require.NotEmpty(t, result)
}
