package weaviate

import (
	"testing"

	"github.com/adrianliechti/llama/test"
)

func TestWeaviate(t *testing.T) {
	context := test.NewContext()

	url := "http://localhost:9084"

	c, err := New(url, "Test", WithEmbedder(context.Embedder))

	if err != nil {
		t.Fatal(err)
	}

	test.TestIndex(t, context, c)
}
