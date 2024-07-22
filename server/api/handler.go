package api

import (
	"encoding/json"
	"mime"
	"net/http"
	"path"

	"github.com/adrianliechti/llama/config"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type Handler struct {
	*config.Config
}

func New(cfg *config.Config) (*Handler, error) {
	h := &Handler{
		Config: cfg,
	}

	return h, nil
}

func (h *Handler) Attach(r chi.Router) {
	r.Get("/index/{index}", h.handleIndexList)
	r.Post("/index/{index}/query", h.handleIndexQuery)

	r.Delete("/index/{index}", h.handleIndexDeletion)

	r.Post("/index/{index}", h.handleIngest)
	r.Post("/index/{index}/{extractor}", h.handleIngestWithExtractor)

	r.Post("/extract/{extractor}", h.handleExtract)
}

func (h *Handler) Handler() http.Handler {
	r := chi.NewRouter()
	h.Attach(r)

	return r
}

func writeJson(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")

	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)

	enc.Encode(v)
}

func writeError(w http.ResponseWriter, code int, err error) {
	w.WriteHeader(code)
	w.Write([]byte(err.Error()))
}

func detectFileName(r *http.Request) string {
	contentType := r.Header.Get("Content-Type")
	contentDisposition := r.Header.Get("Content-Disposition")

	if _, params, err := mime.ParseMediaType(contentDisposition); err == nil {
		if val, ok := params["filename"]; ok && path.Ext(val) != "" {
			return val
		}
	}

	if val, _, err := mime.ParseMediaType(contentType); err == nil {
		if val, ok := typeExtensions[val]; ok {
			return uuid.NewString() + val
		}

		if vals, _ := mime.ExtensionsByType(val); len(vals) > 0 {
			return uuid.NewString() + vals[0]
		}
	}

	return ""
}

var typeExtensions = map[string]string{
	"text/plain": ".txt",
	"text/csv":   ".csv",

	"text/markdown": ".md",
	"text/x-rst":    ".rst",

	"text/rtf":        ".rtf",
	"application/rtf": ".rtf",

	"application/epub+zip": ".epub",

	"message/rfc822":             ".eml",
	"application/vnd.ms-outlook": ".msg",

	"application/msword":            ".doc",
	"application/vnd.ms-excel":      ".xls",
	"application/vnd.ms-powerpoint": ".ppt",

	"application/vnd.oasis.opendocument.text":         ".odt",
	"application/vnd.oasis.opendocument.spreadsheet":  ".ods",
	"application/vnd.oasis.opendocument.presentation": ".odp",

	"application/vnd.openxmlformats-officedocument.wordprocessingml.document":   ".docx",
	"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":         ".xlsx",
	"application/vnd.openxmlformats-officedocument.presentationml.presentation": ".pptx",
}
