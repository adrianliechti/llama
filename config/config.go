package config

import (
	"bytes"
	"os"

	"github.com/adrianliechti/llama/pkg/authorizer"
	"github.com/adrianliechti/llama/pkg/chain"
	"github.com/adrianliechti/llama/pkg/extractor"
	"github.com/adrianliechti/llama/pkg/index"
	"github.com/adrianliechti/llama/pkg/provider"
	"github.com/adrianliechti/llama/pkg/segmenter"
	"github.com/adrianliechti/llama/pkg/summarizer"
	"github.com/adrianliechti/llama/pkg/tool"
	"github.com/adrianliechti/llama/pkg/translator"

	"golang.org/x/time/rate"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Address string

	Authorizers []authorizer.Provider

	models map[string]provider.Model

	completer   map[string]provider.Completer
	embedder    map[string]provider.Embedder
	renderer    map[string]provider.Renderer
	reranker    map[string]provider.Reranker
	synthesizer map[string]provider.Synthesizer
	transcriber map[string]provider.Transcriber

	indexes map[string]index.Provider

	extractors map[string]extractor.Provider
	segmenter  map[string]segmenter.Provider
	summarizer map[string]summarizer.Provider
	translator map[string]translator.Provider

	tools  map[string]tool.Tool
	chains map[string]chain.Provider
}

func Parse(path string) (*Config, error) {
	file, err := parseFile(path)

	if err != nil {
		return nil, err
	}

	c := &Config{
		Address: ":8080",
	}

	if err := c.registerAuthorizer(file); err != nil {
		return nil, err
	}

	if err := c.registerProviders(file); err != nil {
		return nil, err
	}

	if err := c.registerExtractors(file); err != nil {
		return nil, err
	}

	if err := c.registerSegmenters(file); err != nil {
		return nil, err
	}

	if err := c.registerSummarizers(file); err != nil {
		return nil, err
	}

	if err := c.registerTranslators(file); err != nil {
		return nil, err
	}

	if err := c.registerIndexes(file); err != nil {
		return nil, err
	}

	if err := c.registerTools(file); err != nil {
		return nil, err
	}

	if err := c.registerChains(file); err != nil {
		return nil, err
	}

	if err := c.registerRouters(file); err != nil {
		return nil, err
	}

	return c, nil
}

type configFile struct {
	Authorizers []authorizerConfig `yaml:"authorizers"`

	Providers []providerConfig `yaml:"providers"`

	Indexes yaml.Node `yaml:"indexes"`

	Extractors  yaml.Node `yaml:"extractors"`
	Segmenters  yaml.Node `yaml:"segmenters"`
	Translators yaml.Node `yaml:"translators"`

	Tools  yaml.Node `yaml:"tools"`
	Chains yaml.Node `yaml:"chains"`

	Routers yaml.Node `yaml:"routers"`
}

func parseFile(path string) (*configFile, error) {
	data, err := os.ReadFile(path)

	if err != nil {
		return nil, err
	}

	data = []byte(os.ExpandEnv(string(data)))

	var config configFile

	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)

	if err := decoder.Decode(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

func createLimiter(limit *int) *rate.Limiter {
	if limit == nil {
		return nil
	}

	return rate.NewLimiter(rate.Limit(*limit), *limit)
}
