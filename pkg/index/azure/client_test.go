package azure_test

import (
	"os"
	"testing"

	"github.com/adrianliechti/llama/pkg/index/azure"
	"github.com/adrianliechti/llama/test"

	"github.com/stretchr/testify/require"
)

func TestAzure(t *testing.T) {
	context := test.NewContext()

	url := os.Getenv("AZURE_SEARCH_ENDPOINT")
	token := os.Getenv("AZURE_SEARCH_API_KEY")
	index := os.Getenv("AZURE_SEARCH_INDEX_NAME")

	require.NotEmpty(t, url)
	require.NotEmpty(t, token)
	require.NotEmpty(t, index)

	c, err := azure.New(url, index, token)

	if err != nil {
		t.Fatal(err)
	}

	test.TestIndex(t, context, c)
}
