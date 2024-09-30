package duckduckgo_test

import (
	"testing"

	"github.com/adrianliechti/llama/pkg/tool/duckduckgo"
	"github.com/adrianliechti/llama/test"
	"github.com/stretchr/testify/require"
)

func TestSearch(t *testing.T) {
	context := test.NewContext()

	c, err := duckduckgo.New()
	require.NoError(t, err)

	result, err := c.Execute(context.Context, map[string]any{"query": "Meta LLAMA"})
	require.NoError(t, err)

	println(result)
}
