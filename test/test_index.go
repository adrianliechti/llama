package test

import (
	"testing"

	"github.com/adrianliechti/llama/pkg/index"
	"github.com/adrianliechti/llama/pkg/to"
)

func TestIndex(t *testing.T, c *TestContext, i index.Provider) {
	documents := []index.Document{
		{
			ID: "9a7b6b57-a097-492f-b57f-123ae1924f3b",

			Title:    "Embeddings",
			Location: "https://platform.openai.com/docs/guides/embeddings/what-are-embeddings",

			Content: `OpenAI's text embeddings measure the relatedness of text strings. Embeddings are commonly used for:
			Search (where results are ranked by relevance to a query string)
			Clustering (where text strings are grouped by similarity)
			Recommendations (where items with related text strings are recommended)
			Anomaly detection (where outliers with little relatedness are identified)
			Diversity measurement (where similarity distributions are analyzed)
			Classification (where text strings are classified by their most similar label)`,
		},

		{
			ID: "c4f8fd79-5457-4ea8-a567-e9125e445b76",

			Title:    "Text generation models",
			Location: "https://platform.openai.com/docs/guides/text-generation",

			Content: `OpenAI's text generation models (often called generative pre-trained transformers or large language models) have been trained to understand natural language, code, and images. The models provide text outputs in response to their inputs. The text inputs to these models are also referred to as "prompts". Designing a prompt is essentially how you “program” a large language model model, usually by providing instructions or some examples of how to successfully complete a task.
			Using OpenAI's text generation models, you can build applications to:
			- Draft documents
			- Write computer code
			- Answer questions about a knowledge base
			- Analyze texts
			- Give software a natural language interface
			- Tutor in a range of subjects
			- Translate languages
			- Simulate characters for games`,
		},

		{
			ID:      "ec4cf20e-f24d-48d8-88c3-90c9fe7dd435",
			Content: "N/A",
		},
	}

	if err := i.Index(c.Context, documents...); err != nil {
		t.Fatal(err)
	}

	if err := i.Delete(c.Context, "ec4cf20e-f24d-48d8-88c3-90c9fe7dd435"); err != nil {
		t.Fatal(err)
	}

	listed, err := i.List(c.Context, &index.ListOptions{})

	if err != nil {
		t.Fatal(err)
	}

	for _, d := range listed {
		t.Log("documents", d.ID, d.Title, d.Location)
	}

	results, err := i.Query(c.Context, "what are large language models", &index.QueryOptions{Limit: to.Ptr(1)})

	if err != nil {
		t.Fatal(err)
	}

	for _, d := range results {
		t.Log("results", d.Score, d.ID, d.Title, d.Location)
	}
}
