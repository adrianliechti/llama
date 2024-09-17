package llama_test

import (
	"context"
	"testing"

	"github.com/adrianliechti/llama/pkg/provider/llama"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestEmbedder(t *testing.T) {
	ctx := context.Background()

	server, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		Started: true,

		ContainerRequest: testcontainers.ContainerRequest{
			Image: "ghcr.io/ggerganov/llama.cpp:server",

			Cmd: []string{
				"--hf-repo", "nomic-ai/nomic-embed-text-v1.5-GGUF",
				"--hf-file", "nomic-embed-text-v1.5.Q4_K_M.gguf",
				"--embedding",
				"--ctx-size", "8192",
				"--batch-size", "8192",
				"--rope-scaling", "yarn",
				"--rope-freq-scale", ".75",
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

	e, err := llama.NewEmbedder("http://"+url, "default")
	require.NoError(t, err)

	result, err := e.Embed(ctx, "Hallo!")
	require.NoError(t, err)

	require.NotEmpty(t, result.Data)
}
