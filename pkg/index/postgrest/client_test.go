package postgrest_test

import (
	"testing"

	"github.com/adrianliechti/llama/pkg/index/postgrest"
	"github.com/adrianliechti/llama/test"
)

func TestQdrant(t *testing.T) {
	context := test.NewContext()

	url := "localhost:3000"

	c, err := postgrest.New("http://"+url, "docs", postgrest.WithEmbedder(context.Embedder))

	if err != nil {
		t.Fatal(err)
	}

	test.TestIndex(t, context, c)
}
