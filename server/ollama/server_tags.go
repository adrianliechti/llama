package ollama

import (
	"net/http"
	"time"
)

func (s *Server) handleTags(w http.ResponseWriter, r *http.Request) {
	result := &ModelList{}

	timestamp := time.Now().UTC()

	for _, m := range s.Models() {
		result.Models = append(result.Models, Model{
			Name:  m.ID,
			Model: m.ID,

			ModifiedAt: timestamp,

			Details: ModelDetails{
				Format: "gguf",
			},
		})
	}

	writeJson(w, result)
}
