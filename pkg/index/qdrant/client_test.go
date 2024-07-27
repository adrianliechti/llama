package qdrant

import (
	"testing"

	"github.com/adrianliechti/llama/test"
)

func TestQdrant(t *testing.T) {
	context := test.NewContext()

	url := "http://localhost:6333"

	c, err := New(url, "test", WithEmbedder(context.Embedder))

	if err != nil {
		t.Fatal(err)
	}

	test.TestIndex(t, context, c)
}
