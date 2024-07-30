package chroma

import (
	"testing"

	"github.com/adrianliechti/llama/test"
)

func TestChroma(t *testing.T) {
	context := test.NewContext()

	url := "http://localhost:9083"

	c, err := New(url, "test", WithEmbedder(context.Embedder))

	if err != nil {
		t.Fatal(err)
	}

	test.TestIndex(t, context, c)
}
