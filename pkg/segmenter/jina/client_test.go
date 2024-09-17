package jina_test

import (
	"context"
	"strings"
	"testing"

	"github.com/adrianliechti/llama/pkg/segmenter"
	"github.com/adrianliechti/llama/pkg/segmenter/jina"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestExtract(t *testing.T) {
	ctx := context.Background()

	server, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		Started: true,

		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "ghcr.io/adrianliechti/llama-segmenter",
			ExposedPorts: []string{"8000/tcp"},

			WaitingFor: wait.ForLog("Application startup complete"),
		},
	})

	require.NoError(t, err)

	url, err := server.Endpoint(ctx, "")
	require.NoError(t, err)

	s, err := jina.New("http://" + url + "/v1/segment")
	require.NoError(t, err)

	input := segmenter.File{
		Content: strings.NewReader("Hello, World!"),
	}

	segments, err := s.Segment(ctx, input, nil)
	require.NoError(t, err)

	require.NotEmpty(t, segments)
}
