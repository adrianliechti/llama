package memory

import (
	"testing"

	"github.com/adrianliechti/llama/test"
)

func TestMemory(t *testing.T) {
	context := test.NewContext()

	c, err := New(WithEmbedder(context.Embedder))

	if err != nil {
		t.Fatal(err)
	}

	test.TestIndex(t, context, c)
}
