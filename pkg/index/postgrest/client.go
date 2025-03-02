package postgrest

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/adrianliechti/wingman/pkg/index"
)

var _ index.Provider = &Client{}

type Client struct {
	client *http.Client

	url string

	namespace string

	embedder index.Embedder
	reranker index.Reranker
}

func New(url string, namespace string, options ...Option) (*Client, error) {
	c := &Client{
		client: http.DefaultClient,

		url: url,

		namespace: namespace,
	}

	for _, option := range options {
		option(c)
	}

	if c.embedder == nil {
		return nil, errors.New("embedder is required")
	}

	if c.namespace == "" {
		return nil, errors.New("namespace is required")
	}

	return c, nil
}

func jsonReader(v any) io.Reader {
	b := new(bytes.Buffer)

	enc := json.NewEncoder(b)
	enc.SetEscapeHTML(false)

	enc.Encode(v)
	return b
}

func convertError(resp *http.Response) error {
	data, _ := io.ReadAll(resp.Body)

	if len(data) == 0 {
		return errors.New(http.StatusText(resp.StatusCode))
	}

	return errors.New(string(data))
}
