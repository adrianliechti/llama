package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
)

type SegmentService struct {
	Options []RequestOption
}

func NewSegmentService(opts ...RequestOption) *SegmentService {
	return &SegmentService{
		Options: opts,
	}
}

type SegmentRequest struct {
	Content string `json:"content"`

	SegmentLength  *int `json:"segment_length"`
	SegmentOverlap *int `json:"segment_overlap"`
}

type Segment struct {
	Text string `json:"text"`
}

func (r *SegmentService) New(ctx context.Context, body SegmentRequest, opts ...RequestOption) ([]Segment, error) {
	c := newRequestConfig(append(r.Options, opts...)...)

	var data bytes.Buffer

	if err := json.NewEncoder(&data).Encode(body); err != nil {
		return nil, err
	}

	req, _ := http.NewRequestWithContext(ctx, "POST", c.URL+"/v1/segment", &data)
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

	var result struct {
		Segments []Segment `json:"segments,omitempty"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.Segments, nil
}
