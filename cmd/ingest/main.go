package main

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

func main() {
	urlFlag := flag.String("url", "http://localhost:8080", "server url")
	tokenFlag := flag.String("token", "", "server token")
	indexFlag := flag.String("index", "docs", "index name")
	dirFlag := flag.String("dir", ".", "directory")

	flag.Parse()

	url, err := url.Parse(*urlFlag)

	if err != nil {
		panic(err)
	}

	c := client{
		url:    url,
		client: http.DefaultClient,
	}

	ctx := context.Background()

	_ = tokenFlag

	supported := []string{
		".txt", ".md",
	}

	list, err := c.Documents(ctx, *indexFlag)

	if err != nil {
		panic(err)
	}

	revisions := map[string]bool{}
	candidates := map[string]bool{}

	for _, d := range list {
		if revision, ok := d.Metadata["revision"]; ok {
			revisions[revision] = true
		}
	}

	filepath.WalkDir(*dirFlag, func(path string, e fs.DirEntry, err error) error {
		if e.IsDir() {
			return nil
		}

		if !slices.Contains(supported, filepath.Ext(path)) {
			return nil
		}

		data, err := os.ReadFile(path)

		if err != nil {
			return err
		}

		md5_hash := md5.Sum(data)
		md5_text := hex.EncodeToString(md5_hash[:])

		filename := filepath.Base(path)
		filepath, _ := filepath.Rel(*dirFlag, path)

		revision := strings.ToLower(filepath + "@" + md5_text)

		candidates[revision] = true

		if _, ok := revisions[revision]; ok {
			return nil
		}

		content, err := c.Extract(ctx, filename, bytes.NewReader(data), nil)

		if err != nil {
			return err
		}

		if len(content) == 0 {
			return nil
		}

		segments, err := c.Segment(ctx, content, nil)

		if err != nil {
			return err
		}

		var documents []Document

		for i, segment := range segments {
			document := Document{
				Content: segment.Text,

				Metadata: map[string]string{
					"index":    fmt.Sprintf("%d", i),
					"revision": revision,

					"filename": filename,
					"filepath": filepath,
				},
			}

			documents = append(documents, document)
		}

		if err := c.IndexDocuments(ctx, *indexFlag, documents, nil); err != nil {
			return err
		}

		revisions[revision] = true

		return nil
	})

	var deletions []string

	for _, d := range list {
		revision, ok := d.Metadata["revision"]

		if !ok {
			continue
		}

		_, found := candidates[revision]

		if found {
			continue
		}

		deletions = append(deletions, d.ID)
	}

	if len(deletions) > 0 {
		if err := c.DeleteDocuments(ctx, *indexFlag, deletions, nil); err != nil {
			//return err
		}
	}
}

type client struct {
	url    *url.URL
	client *http.Client
}

func (c *client) Extract(ctx context.Context, name string, reader io.Reader, options *ExtractOptions) (string, error) {
	if options == nil {
		options = new(ExtractOptions)
	}

	var body bytes.Buffer
	w := multipart.NewWriter(&body)

	//w.WriteField("model", string(options.Model))
	//w.WriteField("format", string(options.Format))

	file, err := w.CreateFormFile("file", name)

	if err != nil {
		return "", err
	}

	if _, err := io.Copy(file, reader); err != nil {
		return "", err
	}

	w.Close()

	req, _ := http.NewRequestWithContext(ctx, "POST", c.url.JoinPath("/v1/extract").String(), &body)
	req.Header.Set("Content-Type", w.FormDataContentType())

	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", errors.New(resp.Status)
	}

	data, err := io.ReadAll(resp.Body)

	if err != nil {
		return "", err
	}

	return string(data), nil
}

type ExtractOptions struct {
}

func (c *client) Segment(ctx context.Context, content string, options *SegmentOptions) ([]Segment, error) {
	if options == nil {
		options = new(SegmentOptions)
	}

	request := SegmentRequest{
		Content: content,

		SegmentLength:  options.SegmentLength,
		SegmentOverlap: options.SegmentOverlap,
	}

	var body bytes.Buffer

	if err := json.NewEncoder(&body).Encode(request); err != nil {
		return nil, err
	}

	req, _ := http.NewRequestWithContext(ctx, "POST", c.url.JoinPath("/v1/segment").String(), &body)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)

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

type Segment struct {
	Text string `json:"text"`
}

type SegmentOptions struct {
	SegmentLength  *int
	SegmentOverlap *int
}

type SegmentRequest struct {
	Content string `json:"content"`

	SegmentLength  *int `json:"segment_length"`
	SegmentOverlap *int `json:"segment_overlap"`
}

func (c *client) Documents(ctx context.Context, index string) ([]Document, error) {
	req, _ := http.NewRequestWithContext(ctx, "GET", c.url.JoinPath("/v1/index/"+index).String(), nil)

	resp, err := c.client.Do(req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(resp.Status)
	}

	var documents []Document

	if err := json.NewDecoder(resp.Body).Decode(&documents); err != nil {
		return nil, err
	}

	return documents, nil
}

func (c *client) IndexDocuments(ctx context.Context, index string, documents []Document, options *IndexOptions) error {
	if options == nil {
		options = new(IndexOptions)
	}

	var body bytes.Buffer

	if err := json.NewEncoder(&body).Encode(documents); err != nil {
		return err
	}

	req, _ := http.NewRequestWithContext(ctx, "POST", c.url.JoinPath("/v1/index/"+index).String(), &body)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)

	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return errors.New(resp.Status)
	}

	return nil
}

func (c *client) DeleteDocuments(ctx context.Context, index string, ids []string, options any) error {
	var body bytes.Buffer

	if err := json.NewEncoder(&body).Encode(ids); err != nil {
		return err
	}

	req, _ := http.NewRequestWithContext(ctx, "DELETE", c.url.JoinPath("/v1/index/"+index).String(), &body)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)

	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return errors.New(resp.Status)
	}

	return nil
}

type Document struct {
	ID string `json:"id,omitempty"`

	Content string `json:"content,omitempty"`

	Metadata map[string]string `json:"metadata,omitempty"`
}

type IndexOptions struct {
}
