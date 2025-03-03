package duckduckgo_test

import (
	"testing"

	"github.com/adrianliechti/wingman/pkg/index/duckduckgo"
	"github.com/adrianliechti/wingman/test"
	"github.com/stretchr/testify/require"
)

func TestSearch(t *testing.T) {
	context := test.NewContext()

	c, err := duckduckgo.New()
	require.NoError(t, err)

	result, err := c.Query(context.Context, "Meta LLAMA", nil)
	require.NoError(t, err)

	println(result)
}
