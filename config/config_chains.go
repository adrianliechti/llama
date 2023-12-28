package config

import (
	"errors"
	"strings"

	"github.com/adrianliechti/llama/pkg/chain"
	"github.com/adrianliechti/llama/pkg/chain/rag"
)

func createChain(c chainConfig) (chain.Provider, error) {
	switch strings.ToLower(c.Type) {
	case "rag":
		return ragChain(c)

	default:
		return nil, errors.New("invalid chain type: " + c.Type)
	}
}

func ragChain(c chainConfig) (chain.Provider, error) {
	if c.Index == nil {
		return nil, errors.New("missing index configuration")
	}

	index, err := createIndex(*c.Index)

	if err != nil {
		return nil, err
	}

	return rag.New(index, nil, nil), nil
}
