package elasticsearch

import (
	"testing"

	"github.com/adrianliechti/llama/test"
)

func TestElasticsearch(t *testing.T) {
	context := test.NewContext()

	url := "http://localhost:9200"

	c, err := New(url, "test")

	if err != nil {
		t.Fatal(err)
	}

	test.TestIndex(t, context, c)
}
