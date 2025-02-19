package index

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/adrianliechti/llama/pkg/index"
)

func (s *Handler) handleUnstructured(w http.ResponseWriter, r *http.Request) {
	i, err := s.Index(r.PathValue("index"))

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	segmentLength, _ := strconv.Atoi(r.FormValue("segment_length"))
	segmentOverlap, _ := strconv.Atoi(r.FormValue("segment_overlap"))

	file, header, err := r.FormFile("file")

	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	text, err := s.readText(r.Context(), "", header.Filename, file)

	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	segments, err := s.segmentText(r.Context(), "", text, segmentLength, segmentOverlap)

	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	filename := header.Filename
	filepath := header.Filename

	md5_hash := md5.Sum([]byte(text))
	md5_text := hex.EncodeToString(md5_hash[:])

	revision := strings.ToLower(filepath + "@" + md5_text)

	var documents []index.Document

	for i, s := range segments {
		document := index.Document{
			Content: s,

			Metadata: map[string]string{
				"filename": filename,
				"filepath": filepath,

				"revision": revision,

				"index": fmt.Sprintf("%d", i),
			},
		}

		documents = append(documents, document)
	}

	if err := i.Index(r.Context(), documents...); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
