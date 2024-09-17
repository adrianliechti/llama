package config

import (
	"github.com/adrianliechti/llama/pkg/segmenter"
	"github.com/adrianliechti/llama/pkg/segmenter/text"
)

func (cfg *Config) Segmenter(model string) (segmenter.Provider, error) {
	return text.New()
}
