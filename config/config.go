package config

import (
	"bytes"
	"os"

	"github.com/adrianliechti/llama/pkg/authorizer"
	"github.com/adrianliechti/llama/pkg/chain"
	"github.com/adrianliechti/llama/pkg/extractor"
	"github.com/adrianliechti/llama/pkg/index"
	"github.com/adrianliechti/llama/pkg/provider"
	"github.com/adrianliechti/llama/pkg/reranker"
	"github.com/adrianliechti/llama/pkg/summarizer"
	"github.com/adrianliechti/llama/pkg/tool"
	"github.com/adrianliechti/llama/pkg/translator"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Address string

	Authorizers []authorizer.Provider

	models map[string]provider.Model

	completer   map[string]provider.Completer
	embedder    map[string]provider.Embedder
	renderer    map[string]provider.Renderer
	synthesizer map[string]provider.Synthesizer
	transcriber map[string]provider.Transcriber

	reranker   map[string]reranker.Provider
	extractors map[string]extractor.Provider
	summarizer map[string]summarizer.Provider
	translator map[string]translator.Provider

	tools  map[string]tool.Tool
	chains map[string]chain.Provider

	indexes map[string]index.Provider
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

	if err := c.registerIndexes(file); err != nil {
		return nil, err
	}

	if err := c.RegisterExtractors(file); err != nil {
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

	Indexes    map[string]indexConfig     `yaml:"indexes"`
	Extractors map[string]extractorConfig `yaml:"extractors"`

	Routers map[string]routerConfig `yaml:"routers"`

	Tools  map[string]toolConfig  `yaml:"tools"`
	Chains map[string]chainConfig `yaml:"chains"`
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
