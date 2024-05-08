package oai

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"mime"
	"net/http"
	"path"

	"github.com/adrianliechti/llama/pkg/provider"
)

func (s *Server) handleImageGeneration(w http.ResponseWriter, r *http.Request) {
	var req ImageCreateRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	renderer, err := s.Renderer(req.Model)

	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	options := &provider.RenderOptions{}

	image, err := renderer.Render(r.Context(), req.Prompt, options)

	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	data, err := io.ReadAll(image.Content)

	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	result := Image{}

	if req.ResponseFormat == "b64_json" {
		content := base64.StdEncoding.EncodeToString(data)
		result.B64JSON = content
	} else {
		mime := mime.TypeByExtension(path.Ext(image.Name))

		if mime == "" {
			mime = "image/png"
		}

		content := base64.StdEncoding.EncodeToString(data)
		result.URL = "data:" + mime + ";base64," + content
	}

	writeJson(w, result)
}
