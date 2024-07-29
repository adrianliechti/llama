package main

import (
	"context"
	"flag"
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
	extractorFlag := flag.String("extractor", "unstructured", "extractor name")

	flag.Parse()

	ctx := context.Background()

	if pathFlag == nil || *pathFlag == "" {
		wd, err := os.Getwd()

		if err != nil {
			panic(err)
		}

		pathFlag = &wd
	}

	client := http.DefaultClient

	var filetypes []string
	var fileignores []string

	switch strings.ToLower(*extractorFlag) {
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
		panic("unknown extractor")
	}

	url := strings.TrimRight(*urlFlag, "/") + "/api/index/" + *indexFlag + "/" + *extractorFlag

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

		if err != nil {
			slog.Error("failed to open file", "path", path, "error", err)
			return nil
		}

		defer file.Close()

		filename := filepath.Base(path)
		filepath, _ := filepath.Rel(*pathFlag, path)

		req, _ := http.NewRequestWithContext(ctx, "POST", url, file)
		req.Header.Set("Content-Type", "application/octet-stream")
		req.Header.Set("Content-Disposition", "attachment; filename=\""+filename+"\"")

		resp, err := client.Do(req)

		if err != nil {
			slog.Error("failed to index document", "path", filepath, "error", err)
			return nil
		}

		defer resp.Body.Close()

		if !(resp.StatusCode == 200 || resp.StatusCode == 204) {
			slog.Error("failed to index document", "path", filepath, "status", resp.Status)
			return nil
		}

		slog.Info("indexed", "path", filepath)

		return nil
	})

	if err != nil {
		panic(err)
	}
}
