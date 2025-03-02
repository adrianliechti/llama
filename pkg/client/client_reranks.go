package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/adrianliechti/wingman/server/api"
)

type RerankService struct {
	Options []RequestOption
}

func NewRerankService(opts ...RequestOption) *RerankService {
	return &RerankService{
		Options: opts,
	}
}

type Rerank = api.Result
type RerankRequest = api.RerankRequest

func (r *RerankService) New(ctx context.Context, input RerankRequest, opts ...RequestOption) ([]Rerank, error) {
	c := newRequestConfig(append(r.Options, opts...)...)

	var data bytes.Buffer

	if err := json.NewEncoder(&data).Encode(input); err != nil {
		return nil, err
	}

	req, _ := http.NewRequestWithContext(ctx, "POST", c.URL+"/v1/rerank", &data)
	req.Header.Set("Content-Type", "application/json")

	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}

	resp, err := c.Client.Do(req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(resp.Status)
	}

	var result api.RerankResponse

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.Results, nil
}
