package weaviate_test

import (
	"testing"

	"github.com/adrianliechti/wingman/pkg/index/weaviate"
	"github.com/adrianliechti/wingman/test"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestWeaviate(t *testing.T) {
	context := test.NewContext()

	server, err := testcontainers.GenericContainer(context.Context, testcontainers.GenericContainerRequest{
		Started: true,

		ContainerRequest: testcontainers.ContainerRequest{
			Image: "cr.weaviate.io/semitechnologies/weaviate:1.26.4",
			Env: map[string]string{
				"CLUSTER_HOSTNAME":          "node1",
				"DEFAULT_VECTORIZER_MODULE": "none",
			},
			ExposedPorts: []string{"8080/tcp"},
			WaitingFor:   wait.ForLog("node reporting ready"),
		},
	})

	require.NoError(t, err)

	url, err := server.Endpoint(context.Context, "")
	require.NoError(t, err)

	c, err := weaviate.New("http://"+url, "Test", weaviate.WithEmbedder(context.Embedder))
	require.NoError(t, err)

	test.TestIndex(t, context, c)
}
