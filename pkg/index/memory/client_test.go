package memory_test

import (
	"testing"

	"github.com/adrianliechti/wingman/pkg/index/memory"
	"github.com/adrianliechti/wingman/test"

	"github.com/stretchr/testify/require"
)

func TestMemory(t *testing.T) {
	context := test.NewContext()

	c, err := memory.New(memory.WithEmbedder(context.Embedder))
	require.NoError(t, err)

	test.TestIndex(t, context, c)
}
