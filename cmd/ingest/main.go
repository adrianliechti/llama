package main

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/adrianliechti/llama/pkg/client"
	"github.com/adrianliechti/llama/pkg/to"

	"github.com/openai/openai-go"
)

var (
	urlFlag   = flag.String("url", "http://localhost:8080", "platform url")
	tokenFlag = flag.String("token", "", "platform token")

	indexFlag = flag.String("index", "docs", "index name")
	dirFlag   = flag.String("dir", ".", "index directory")

	embeddingFlag = flag.String("embedding", "", "embedding model")
)

func main() {
	flag.Parse()

	ctx := context.Background()

	opts := []client.RequestOption{}

	if *tokenFlag != "" {
		opts = append(opts, client.WithToken(*tokenFlag))
	}

	c := client.New(*urlFlag, opts...)

	if err := IndexDir(ctx, c, *indexFlag, *dirFlag); err != nil {
		panic(err)
	}
}

func IndexDir(ctx context.Context, c *client.Client, index, root string) error {
	supported := []string{
		".csv",
		".md",
		".rst",
		".tsv",
		".txt",

		".pdf",

		// ".jpg", ".jpeg",
		// ".png",
		// ".bmp",
		// ".tiff",
		// ".heif",

		".docx",
		".pptx",
		".xlsx",
	}

	var result error

	// list, err := c.Documents(ctx, index)

	// if err != nil {
	// 	return err
	// }

	// revisions := map[string]bool{}
	// candidates := map[string]bool{}

	// for _, d := range list {
	// 	if revision, ok := d.Metadata["revision"]; ok {
	// 		revisions[revision] = true
	// 	}
	// }

	revisions := map[string]string{}

	filepath.WalkDir(root, func(path string, e fs.DirEntry, err error) error {
		if err != nil {
			result = errors.Join(result, err)
			return nil
		}

		if strings.Contains(path, ".cache") {
			return nil
		}

		if e.IsDir() || !slices.Contains(supported, filepath.Ext(path)) {
			return nil
		}

		data, err := os.ReadFile(path)

		if err != nil {
			result = errors.Join(result, err)
			return nil
		}

		md5_hash := md5.Sum(data)
		md5_text := hex.EncodeToString(md5_hash[:])

		cachedir := filepath.Join(root, ".cache", md5_text[0:2], md5_text[2:4], md5_text)
		os.MkdirAll(cachedir, 0755)

		info, err := e.Info()

		if err != nil {
			result = errors.Join(result, err)
			return nil
		}

		rel, _ := filepath.Rel(root, path)

		name := filepath.Base(path)
		title := strings.TrimSuffix(name, filepath.Ext(name))

		metadata := Metadata{
			Name: filepath.Base(path),
			Path: "/" + rel,

			Title: title,

			MD5: md5_text,

			Size: info.Size(),
			Time: info.ModTime(),
		}

		if !exists(cachedir, "metadata.json") {
			if err := writeJSON(cachedir, "metadata.json", metadata); err != nil {
				result = errors.Join(result, err)
				return nil
			}
		}

		if !exists(cachedir, "content.txt") {
			body := client.ExtractionRequest{
				Name:   metadata.Name,
				Reader: bytes.NewReader(data),
			}

			content, err := c.Extractions.New(ctx, body)

			if err != nil {
				result = errors.Join(result, err)
				return nil
			}

			if err := writeData(cachedir, "content.txt", []byte(content.Text)); err != nil {
				result = errors.Join(result, err)
				return nil
			}
		}

		if *embeddingFlag != "" && !exists(cachedir, "embeddings.json") {
			text, err := readText(cachedir, "content.txt")

			if err != nil {
				result = errors.Join(result, err)
				return nil
			}

			segments, err := c.Segments.New(ctx, client.SegmentRequest{
				Text: text,

				SegmentLength:  to.Ptr(3000),
				SegmentOverlap: to.Ptr(1500),
			})

			if err != nil {
				result = errors.Join(result, err)
				return nil
			}

			embeddings := Embeddings{
				Model: *embeddingFlag,
			}

			titleEmbedding, err := c.Embeddings.New(ctx, openai.EmbeddingNewParams{
				Model: openai.F(*embeddingFlag),
				Input: openai.F[openai.EmbeddingNewParamsInputUnion](openai.EmbeddingNewParamsInputArrayOfStrings([]string{title})),
			})

			if err != nil {
				result = errors.Join(result, err)
				return nil
			}

			embeddings.Segments = append(embeddings.Segments, Segment{
				Text:      title,
				Embedding: titleEmbedding.Data[0].Embedding,
			})

			for _, segment := range segments {
				segmentEmbedding, err := c.Embeddings.New(ctx, openai.EmbeddingNewParams{
					Model: openai.F(*embeddingFlag),
					Input: openai.F[openai.EmbeddingNewParamsInputUnion](openai.EmbeddingNewParamsInputArrayOfStrings([]string{segment.Text})),
				})

				if err != nil {
					result = errors.Join(result, err)
					return nil
				}

				embeddings.Segments = append(embeddings.Segments, Segment{
					Text:      segment.Text,
					Embedding: segmentEmbedding.Data[0].Embedding,
				})
			}

			if err := writeJSON(cachedir, "embeddings.json", embeddings); err != nil {
				result = errors.Join(result, err)
				return nil
			}
		}

		revisions[metadata.Path] = metadata.MD5

		println(metadata.Path)

		// 	filename := filepath.Base(path)
		// 	filepath, _ := filepath.Rel(root, path)

		// 	revision := strings.ToLower(filepath + "@" + md5_text)

		// 	candidates[revision] = true

		// 	if _, ok := revisions[revision]; ok {
		// 		return nil
		// 	}

		// 	fmt.Printf("Indexing %s...\n", path)

		// 	content, err := c.Extract(ctx, filename, bytes.NewReader(data), nil)

		// 	if err != nil {
		// 		result = errors.Join(result, err)
		// 		return err
		// 	}

		// 	if len(content) == 0 {
		// 		return nil
		// 	}

		// 	segments, err := c.Segment(ctx, content, nil)

		// 	if err != nil {
		// 		result = errors.Join(result, err)
		// 		return err
		// 	}

		// 	var documents []Document

		// 	for i, segment := range segments {
		// 		document := Document{
		// 			Content: segment.Text,

		// 			Metadata: map[string]string{
		// 				"filename": filename,
		// 				"filepath": filepath,

		// 				"revision": revision,

		// 				"index": fmt.Sprintf("%d", i),
		// 			},
		// 		}

		// 		documents = append(documents, document)
		// 	}

		// 	if err := c.IndexDocuments(ctx, index, documents, nil); err != nil {
		// 		result = errors.Join(result, err)
		// 		return err
		// 	}

		// 	revisions[revision] = true

		// 	return nil
		// })

		// var deletions []string

		// for _, d := range list {
		// 	revision, ok := d.Metadata["revision"]

		// 	if !ok {
		// 		continue
		// 	}

		// 	_, found := candidates[revision]

		// 	if found {
		// 		continue
		// 	}

		// 	deletions = append(deletions, d.ID)
		// }

		// if len(deletions) > 0 {
		// 	if err := c.DeleteDocuments(ctx, index, deletions, nil); err != nil {
		// 		result = errors.Join(result, err)
		// 	}

		return nil
	})

	return result
}

type Metadata struct {
	Name string `json:"name"`
	Path string `json:"path"`

	Title string `json:"title"`

	MD5 string `json:"md5"`

	Size int64     `json:"size"`
	Time time.Time `json:"time"`
}

type Embeddings struct {
	Model string `json:"model"`

	Segments []Segment `json:"segments"`
}

type Segment struct {
	Text string `json:"text"`

	Embedding []float64 `json:"embedding"`
}

func exists(path, name string) bool {
	info, err := os.Stat(filepath.Join(path, name))

	if err != nil {
		if os.IsNotExist(err) {
			return false
		}

		return false
	}

	return !info.IsDir()
}

func readData(dir, name string) ([]byte, error) {
	return os.ReadFile(filepath.Join(dir, name))
}

func readText(dir, name string) (string, error) {
	data, err := readData(dir, name)

	if err != nil {
		return "", err
	}

	return string(data), nil
}

func readJSON(dir, name string, v any) error {
	data, err := readData(dir, name)

	if err != nil {
		return err
	}

	return json.Unmarshal(data, v)
}

func writeData(dir, name string, data []byte) error {
	return os.WriteFile(filepath.Join(dir, name), data, 0644)
}

func writeJSON(dir, name string, v any) error {
	data, err := json.MarshalIndent(v, "", "  ")

	if err != nil {
		return err
	}

	return writeData(dir, name, data)
}

// type client struct {
// 	url    *url.URL
// 	token  string
// 	client *http.Client
// }

// func (c *client) Extract(ctx context.Context, name string, reader io.Reader, options *ExtractOptions) (string, error) {
// 	if options == nil {
// 		options = new(ExtractOptions)
// 	}

// 	var body bytes.Buffer
// 	w := multipart.NewWriter(&body)

// 	//w.WriteField("model", string(options.Model))
// 	//w.WriteField("format", string(options.Format))

// 	file, err := w.CreateFormFile("file", name)

// 	if err != nil {
// 		return "", err
// 	}

// 	if _, err := io.Copy(file, reader); err != nil {
// 		return "", err
// 	}

// 	w.Close()

// 	req, _ := http.NewRequestWithContext(ctx, "POST", c.url.JoinPath("/v1/extract").String(), &body)
// 	req.Header.Set("Content-Type", w.FormDataContentType())

// 	resp, err := http.DefaultClient.Do(req)

// 	if err != nil {
// 		return "", err
// 	}

// 	defer resp.Body.Close()

// 	if resp.StatusCode != http.StatusOK {
// 		return "", errors.New(resp.Status)
// 	}

// 	data, err := io.ReadAll(resp.Body)

// 	if err != nil {
// 		return "", err
// 	}

// 	return string(data), nil
// }

// type ExtractOptions struct {
// }

// func (c *client) Segment(ctx context.Context, content string, options *SegmentOptions) ([]Segment, error) {
// 	if options == nil {
// 		options = new(SegmentOptions)
// 	}

// 	request := SegmentRequest{
// 		Content: content,

// 		SegmentLength:  options.SegmentLength,
// 		SegmentOverlap: options.SegmentOverlap,
// 	}

// 	var body bytes.Buffer

// 	if err := json.NewEncoder(&body).Encode(request); err != nil {
// 		return nil, err
// 	}

// 	req, _ := http.NewRequestWithContext(ctx, "POST", c.url.JoinPath("/v1/segment").String(), &body)
// 	req.Header.Set("Content-Type", "application/json")

// 	if c.token != "" {
// 		req.Header.Set("Authorization", "Bearer "+c.token)
// 	}

// 	resp, err := c.client.Do(req)

// 	if err != nil {
// 		return nil, err
// 	}

// 	defer resp.Body.Close()

// 	if resp.StatusCode != http.StatusOK {
// 		return nil, errors.New(resp.Status)
// 	}

// 	var result struct {
// 		Segments []Segment `json:"segments,omitempty"`
// 	}

// 	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
// 		return nil, err
// 	}

// 	return result.Segments, nil
// }

// type Segment struct {
// 	Text string `json:"text"`
// }

// type SegmentOptions struct {
// 	SegmentLength  *int
// 	SegmentOverlap *int
// }

// type SegmentRequest struct {
// 	Content string `json:"content"`

// 	SegmentLength  *int `json:"segment_length"`
// 	SegmentOverlap *int `json:"segment_overlap"`
// }

// func (c *client) Documents(ctx context.Context, index string) ([]Document, error) {
// 	req, _ := http.NewRequestWithContext(ctx, "GET", c.url.JoinPath("/v1/index/"+index).String(), nil)

// 	if c.token != "" {
// 		req.Header.Set("Authorization", "Bearer "+c.token)
// 	}

// 	resp, err := c.client.Do(req)

// 	if err != nil {
// 		return nil, err
// 	}

// 	defer resp.Body.Close()

// 	if resp.StatusCode != http.StatusOK {
// 		return nil, errors.New(resp.Status)
// 	}

// 	var documents []Document

// 	if err := json.NewDecoder(resp.Body).Decode(&documents); err != nil {
// 		return nil, err
// 	}

// 	return documents, nil
// }

// func (c *client) IndexDocuments(ctx context.Context, index string, documents []Document, options *IndexOptions) error {
// 	if options == nil {
// 		options = new(IndexOptions)
// 	}

// 	var body bytes.Buffer

// 	if err := json.NewEncoder(&body).Encode(documents); err != nil {
// 		return err
// 	}

// 	req, _ := http.NewRequestWithContext(ctx, "POST", c.url.JoinPath("/v1/index/"+index).String(), &body)
// 	req.Header.Set("Content-Type", "application/json")

// 	if c.token != "" {
// 		req.Header.Set("Authorization", "Bearer "+c.token)
// 	}

// 	resp, err := c.client.Do(req)

// 	if err != nil {
// 		return err
// 	}

// 	defer resp.Body.Close()

// 	if resp.StatusCode != http.StatusNoContent {
// 		return errors.New(resp.Status)
// 	}

// 	return nil
// }

// func (c *client) DeleteDocuments(ctx context.Context, index string, ids []string, options any) error {
// 	var body bytes.Buffer

// 	if err := json.NewEncoder(&body).Encode(ids); err != nil {
// 		return err
// 	}

// 	req, _ := http.NewRequestWithContext(ctx, "DELETE", c.url.JoinPath("/v1/index/"+index).String(), &body)
// 	req.Header.Set("Content-Type", "application/json")

// 	if c.token != "" {
// 		req.Header.Set("Authorization", "Bearer "+c.token)
// 	}

// 	resp, err := c.client.Do(req)

// 	if err != nil {
// 		return err
// 	}

// 	defer resp.Body.Close()

// 	if resp.StatusCode != http.StatusNoContent {
// 		return errors.New(resp.Status)
// 	}

// 	return nil
// }

// type Document struct {
// 	ID string `json:"id,omitempty"`

// 	Content string `json:"content,omitempty"`

// 	Metadata map[string]string `json:"metadata,omitempty"`
// }

// type IndexOptions struct {
// }
