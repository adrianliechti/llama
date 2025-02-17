package unstructured_test

import (
	"context"
	"strings"
	"testing"

	"github.com/adrianliechti/llama/pkg/segmenter/unstructured"

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

	s, err := unstructured.New("http://" + url)
	require.NoError(t, err)

	input := strings.Repeat("Hello, World! ", 2000)

	segments, err := s.Segment(ctx, input, nil)
	require.NoError(t, err)

	require.NotEmpty(t, segments)
}
