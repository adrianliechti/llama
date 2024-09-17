package elasticsearch_test

import (
	"testing"

	"github.com/adrianliechti/llama/pkg/index/elasticsearch"
	"github.com/adrianliechti/llama/test"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestElasticsearch(t *testing.T) {
	context := test.NewContext()

	server, err := testcontainers.GenericContainer(context.Context, testcontainers.GenericContainerRequest{
		Started: true,

		ContainerRequest: testcontainers.ContainerRequest{
			Image: "docker.elastic.co/elasticsearch/elasticsearch:8.15.1",
			Env: map[string]string{
				"ES_JAVA_OPTS":           "-Xms1g -Xmx1g",
				"discovery.type":         "single-node",
				"xpack.security.enabled": "false",
				"node.name":              "test",
				"cluster.name":           "test",
			},
			ExposedPorts: []string{"9200/tcp"},
			WaitingFor:   wait.ForExposedPort(),
		},
	})

	require.NoError(t, err)

	url, err := server.Endpoint(context.Context, "")
	require.NoError(t, err)

	c, err := elasticsearch.New("http://"+url, "test")
	require.NoError(t, err)

	test.TestIndex(t, context, c)
}
