package main

import (
	"github.com/adrianliechti/llama/config"
	"github.com/adrianliechti/llama/pkg/server"
)

func main() {
	cfg, err := config.FromEnvironment()

	if err != nil {
		panic(err)
	}

	s := server.New(cfg)
	s.ListenAndServe()
}
