package memory_test

import (
	"testing"

	"github.com/adrianliechti/llama/pkg/index/memory"
	"github.com/adrianliechti/llama/test"

	"github.com/stretchr/testify/require"
)

func TestMemory(t *testing.T) {
	context := test.NewContext()

	c, err := memory.New(memory.WithEmbedder(context.Embedder))
	require.NoError(t, err)

	test.TestIndex(t, context, c)
}
