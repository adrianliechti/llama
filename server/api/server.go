package api

import (
	"encoding/json"
	"mime"
	"net/http"

	"github.com/adrianliechti/llama/config"
	"github.com/google/uuid"

	"github.com/go-chi/chi/v5"
)

type Server struct {
	*config.Config
	http.Handler
}

func New(cfg *config.Config) (*Server, error) {
	r := chi.NewRouter()

	s := &Server{
		Config:  cfg,
		Handler: r,
	}

	r.Post("/extract/{extracter}", s.handleExtract)

	r.Get("/index/{index}", s.handleIndexList)
	r.Post("/index/{index}", s.handleIndexIngest)
	r.Post("/index/{index}/query", s.handleIndexQuery)
	r.Post("/index/{index}/{extracter}", s.handleIndexWithExtracter)

	return s, nil
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
	var name string

	contentType := r.Header.Get("Content-Type")
	contentDisposition := r.Header.Get("Content-Disposition")

	if val, _, err := mime.ParseMediaType(contentType); err == nil {
		if vals, _ := mime.ExtensionsByType(val); err == nil && len(vals) > 0 {
			name = uuid.NewString() + vals[0]
		}
	}

	if _, params, err := mime.ParseMediaType(contentDisposition); err == nil {
		if val := params["filename"]; val != "" {
			name = val
		}
	}

	return name
}
