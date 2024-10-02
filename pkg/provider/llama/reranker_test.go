package llama_test

import (
	"context"
	"testing"

	"github.com/adrianliechti/llama/pkg/provider/llama"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestReranker(t *testing.T) {
	ctx := context.Background()

	server, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		Started: true,

		ContainerRequest: testcontainers.ContainerRequest{
			Image: "ghcr.io/ggerganov/llama.cpp:server",

			Cmd: []string{
				"--hf-repo", "gpustack/bge-reranker-v2-m3-GGUF",
				"--hf-file", "bge-reranker-v2-m3-Q4_K_M.gguf",
				"--reranking",
				"--ctx-size", "8192",
			},

			Mounts: testcontainers.Mounts(
				testcontainers.ContainerMount{
					Target: "/root/.cache/llama.cpp",
					Source: testcontainers.DockerVolumeMountSource{
						Name: "llama",
					},
				},
			),

			ExposedPorts: []string{"8080/tcp"},

			WaitingFor: wait.ForLog("starting the main loop"),
		},
	})

	require.NoError(t, err)

	url, err := server.Endpoint(ctx, "")
	require.NoError(t, err)

	e, err := llama.NewReranker("http://"+url, "default")
	require.NoError(t, err)

	result, err := e.Rerank(ctx, "Hallo!", []string{
		"hi",
		"it is a bear",
		"The giant panda (Ailuropoda melanoleuca), sometimes called a panda bear or simply panda, is a bear species endemic to China.",
	}, nil)
	require.NoError(t, err)

	require.NotEmpty(t, result)
}
