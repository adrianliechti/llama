package jina

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/adrianliechti/wingman/pkg/segmenter"
)

var _ segmenter.Provider = &Client{}

type Client struct {
	client *http.Client

	url   string
	token string
}

func New(url string, options ...Option) (*Client, error) {
	if url == "" {
		url = "https://segment.jina.ai/"
	}

	c := &Client{
		client: http.DefaultClient,

		url: url,
	}

	for _, option := range options {
		option(c)
	}

	return c, nil
}

func (c *Client) Segment(ctx context.Context, input string, options *segmenter.SegmentOptions) ([]segmenter.Segment, error) {
	if options == nil {
		options = new(segmenter.SegmentOptions)
	}

	body := SegmentRequest{
		Content: input,

		ReturnChunks:   true,
		MaxChunkLength: 1000,
	}

	if options.SegmentLength != nil {
		body.MaxChunkLength = *options.SegmentLength
	}

	req, _ := http.NewRequestWithContext(ctx, "POST", c.url, jsonReader(body))
	req.Header.Set("Content-Type", "application/json")

	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.client.Do(req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, convertError(resp)
	}

	var result SegmentResponse

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	var segments []segmenter.Segment

	for _, chunk := range result.Chunks {
		segment := segmenter.Segment{
			Text: chunk,
		}

		segments = append(segments, segment)
	}

	return segments, nil
}

func convertError(resp *http.Response) error {
	data, _ := io.ReadAll(resp.Body)

	if len(data) == 0 {
		return errors.New(http.StatusText(resp.StatusCode))
	}

	return errors.New(string(data))
}

func jsonReader(v any) io.Reader {
	b := new(bytes.Buffer)

	enc := json.NewEncoder(b)
	enc.SetEscapeHTML(false)

	enc.Encode(v)
	return b
}
