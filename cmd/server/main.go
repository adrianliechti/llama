package main

import (
	"flag"
	"fmt"

	"github.com/adrianliechti/llama/config"
	"github.com/adrianliechti/llama/server"
)

func main() {
	portFlag := flag.Int("port", 8080, "server port")
	addressFlag := flag.String("address", "", "server address")
	configFlag := flag.String("config", "config.yaml", "configuration path")

	flag.Parse()

	cfg, err := config.Parse(*configFlag)

	if err != nil {
		panic(err)
	}

	cfg.Address = fmt.Sprintf("%s:%d", *addressFlag, *portFlag)

	s, err := server.New(cfg)

	if err != nil {
		panic(err)
	}

	if err := s.ListenAndServe(); err != nil {
		panic(err)
	}
}
