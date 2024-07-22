package openai

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"mime"
	"net/http"
	"path"

	"github.com/adrianliechti/llama/pkg/provider"
)

func (h *Handler) handleImageGeneration(w http.ResponseWriter, r *http.Request) {
	var req ImageCreateRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	renderer, err := h.Renderer(req.Model)

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

	result := ImageList{}

	if req.ResponseFormat == "b64_json" {
		result.Images = []Image{
			{
				B64JSON: base64.StdEncoding.EncodeToString(data),
			},
		}

	} else {
		mime := mime.TypeByExtension(path.Ext(image.Name))

		if mime == "" {
			mime = "image/png"
		}

		result.Images = []Image{
			{
				URL: "data:" + mime + ";base64," + base64.StdEncoding.EncodeToString(data),
			},
		}
	}

	writeJson(w, result)
}
