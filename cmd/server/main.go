package main

import (
	"github.com/adrianliechti/llama/config"
	"github.com/adrianliechti/llama/pkg/server"
)

func main() {
	cfg, err := config.Parse("")

	if err != nil {
		panic(err)
	}

	s, err := server.New(cfg)

	if err != nil {
		panic(err)
	}

	if err := s.ListenAndServe(); err != nil {
		panic(err)
	}
}
