package oai

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func (s *Server) handleFiles(w http.ResponseWriter, r *http.Request) {
	purpose := "assistants"

	id := "file-id-1"

	filename := "Test.txt"
	bytes := int64(19)

	file1 := File{
		Object: "file",

		ID: id,

		Purpose:   purpose,
		CreatedAt: time.Now().Unix(),

		Filename: filename,
		Bytes:    bytes,
	}

	list := FileList{
		Object: "list",

		Data: []File{
			file1,
		},
	}

	writeJson(w, list)
}

func (s *Server) handleFile(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	purpose := "assistants"

	filename := "Test.txt"
	bytes := int64(19)

	result := File{
		Object: "file",

		ID: id,

		Purpose:   purpose,
		CreatedAt: time.Now().Unix(),

		Filename: filename,
		Bytes:    bytes,
	}

	writeJson(w, result)
}

func (s *Server) handleFileContent(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	_ = id

	content := "This is a test file"

	w.Write([]byte(content))
}

func (s *Server) handleCreateFile(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(16 << 20); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	purpose := r.FormValue("purpose")

	file, fileHeader, err := r.FormFile("file")

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	defer file.Close()

	result := File{
		Object: "file",

		ID: uuid.NewString(),

		Purpose:   purpose,
		CreatedAt: time.Now().Unix(),

		Filename: fileHeader.Filename,
		Bytes:    fileHeader.Size,
	}

	writeJson(w, result)
}

func (s *Server) handleDeleteFile(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	_ = id

	w.WriteHeader(http.StatusNoContent)
}
