package server

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"chat/pkg/auth"
	"chat/provider"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/sashabaranov/go-openai"
)

type Server struct {
	router *chi.Mux

	auth     auth.Provider
	provider provider.Provider
}

func New(a auth.Provider, p provider.Provider) *Server {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	s := &Server{
		router: r,

		auth:     a,
		provider: p,
	}

	r.Use(s.handleAuth)

	r.Get("/v1/models", s.handleModels)
	r.Get("/v1/model/{id}", s.handleModel)

	r.Post("/v1/chat/completions", s.handleChatCompletions)

	return s
}

func (s *Server) ListenAndServe(addr string) error {
	return http.ListenAndServe(addr, s.router)
}

func (s *Server) handleAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		if err := s.auth.Verify(ctx, r); err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (s *Server) handleModels(w http.ResponseWriter, r *http.Request) {
	models, err := s.provider.Models(r.Context())

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	result := openai.ModelsList{
		Models: models,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (s *Server) handleModel(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	models, err := s.provider.Models(r.Context())

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	for _, m := range models {
		if !strings.EqualFold(id, m.ID) {
			continue
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(m)
		return
	}

	w.WriteHeader(http.StatusNotFound)
}

func (s *Server) handleChatCompletions(w http.ResponseWriter, r *http.Request) {
	var req openai.ChatCompletionRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if req.Stream {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		done := make(chan error)
		stream := make(chan openai.ChatCompletionStreamResponse)

		// defer func() {
		// 	close(done)
		// 	close(stream)
		// }()

		go func() {
			done <- s.provider.ChatStream(r.Context(), req, stream)
		}()

		for {
			select {
			case err := <-done:
				fmt.Fprintf(w, "data: [DONE]\n\n")
				w.(http.Flusher).Flush()

				if err != nil {
					slog.Error("error in chat completion", "error", err)
				}

				return

			case resp := <-stream:
				data, _ := json.Marshal(resp)

				fmt.Fprintf(w, "data: %s\n\n", string(data))
				w.(http.Flusher).Flush()

			case <-r.Context().Done():
				return
			}
		}
	} else {
		result, err := s.provider.Chat(r.Context(), req)

		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	}
}
