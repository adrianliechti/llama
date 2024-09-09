package main

import (
	"context"
	"errors"
	"flag"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

func main() {
	urlFlag := flag.String("url", "http://localhost:8080", "server url")
	pathFlag := flag.String("path", "", "documents path")
	indexFlag := flag.String("index", "docs", "index name")
	partitionerFlag := flag.String("partitioner", "unstructured", "partitioner name")

	flag.Parse()

	ctx := context.Background()

	if pathFlag == nil || *pathFlag == "" {
		wd, err := os.Getwd()

		if err != nil {
			panic(err)
		}

		pathFlag = &wd
	}

	fi, err := os.Stat(*pathFlag)

	if err != nil {
		panic(err)
	}

	var filetypes []string
	var fileignores []string

	switch strings.ToLower(*partitionerFlag) {
	case "text":
		filetypes = []string{
			".txt", ".html", ".md",
		}

	case "code":
		filetypes = []string{
			".c", ".h", ".cpp", ".hpp", ".m", ".cs", ".vb", ".java", ".js", ".mjs", ".py", ".rb", ".sql", ".sh", ".bat",
			".swift", ".kt", ".kts", ".go", ".rs", ".ts", ".tsx", ".scala", ".pl", ".pm", ".lua", ".dart", ".groovy", ".gvy", ".jl",
		}

		fileignores = []string{
			".pb.go", ".generated.go",
			".min.js", "d.ts", "node_modules/",
		}

	case "unstructured":
		filetypes = []string{
			".txt", ".eml", ".msg", ".html", ".md", ".rst", ".rtf",
			".jpeg", ".png",
			".doc", ".docx", ".ppt", ".pptx", ".pdf", ".odt", ".epub", ".csv", ".tsv", ".xlsx",
		}
	default:
		panic("unknown partitioner")
	}

	if fi.IsDir() {
		err := filepath.Walk(*pathFlag, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if strings.Contains(path, "/.") {
				return nil
			}

			if !slices.Contains(filetypes, filepath.Ext(path)) {
				return nil
			}

			if slices.ContainsFunc(fileignores, func(s string) bool {
				return strings.Contains(strings.ToLower(path), strings.ToLower(s))
			}) {
				return nil
			}

			file, err := os.Open(path)

			filename := filepath.Base(path)
			filepath, _ := filepath.Rel(*pathFlag, path)

			if err != nil {
				slog.Error("failed to open file", "path", filepath, "error", err)
				return nil
			}

			defer file.Close()

			if err := ingestDocument(ctx, *urlFlag, *indexFlag, *partitionerFlag, filename, file); err != nil {
				slog.Error("failed to ingest file", "path", filepath, "error", err)
				return nil
			}

			slog.Info("indexed", "path", filepath)

			return nil
		})

		if err != nil {
			panic(err)
		}
	} else {
		file, err := os.Open(*pathFlag)

		filename := filepath.Base(*pathFlag)
		filepath := *pathFlag

		if err != nil {
			slog.Error("failed to open file", "path", filepath, "error", err)
			panic(err)
		}

		defer file.Close()

		if err := ingestDocument(ctx, *urlFlag, *indexFlag, *partitionerFlag, filename, file); err != nil {
			panic(err)
		}

		slog.Info("indexed", "path", filepath)
	}
}

func ingestDocument(ctx context.Context, baseURL, index, partitioner, filename string, file io.Reader) error {
	client := http.DefaultClient

	url := strings.TrimRight(baseURL, "/") + "/api/index/" + index + "/" + partitioner

	req, _ := http.NewRequestWithContext(ctx, "POST", url, file)
	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Set("Content-Disposition", "attachment; filename=\""+filename+"\"")

	resp, err := client.Do(req)

	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if !(resp.StatusCode == 200 || resp.StatusCode == 204) {
		return errors.New("failed to index document")
	}

	return nil
}
