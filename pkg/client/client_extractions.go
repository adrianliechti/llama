package client

import (
	"bytes"
	"context"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
)

type ExtractionService struct {
	Options []RequestOption
}

func NewExtractionService(opts ...RequestOption) *ExtractionService {
	return &ExtractionService{
		Options: opts,
	}
}

type ExtractionNewParams struct {
	Name   string
	Reader io.Reader
}

type Extraction struct {
	Text string `json:"text"`
}

func (r *ExtractionService) New(ctx context.Context, body ExtractionNewParams, opts ...RequestOption) (*Extraction, error) {
	c := newRequestConfig(append(r.Options, opts...)...)

	var data bytes.Buffer
	w := multipart.NewWriter(&data)

	//w.WriteField("model", string(options.Model))
	//w.WriteField("format", string(options.Format))

	file, err := w.CreateFormFile("file", body.Name)

	if err != nil {
		return nil, err
	}

	if _, err := io.Copy(file, body.Reader); err != nil {
		return nil, err
	}

	w.Close()

	req, _ := http.NewRequestWithContext(ctx, "POST", c.URL+"/v1/extract", &data)
	req.Header.Set("Content-Type", w.FormDataContentType())

	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}

	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(resp.Status)
	}

	result, err := io.ReadAll(resp.Body)

	if err != nil {
		return nil, err
	}

	return &Extraction{
		Text: string(result),
	}, nil
}
