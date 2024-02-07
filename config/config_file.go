package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

func parseFile(path string) (*configFile, error) {
	data, err := os.ReadFile(path)

	if err != nil {
		return nil, err
	}

	var config configFile

	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

type configFile struct {
	Authorizers []authorizerConfig `yaml:"authorizers"`

	Providers []providerConfig `yaml:"providers"`

	Indexes     map[string]indexConfig      `yaml:"indexes"`
	Extracters  map[string]extracterConfig  `yaml:"extracters"`
	Classifiers map[string]classifierConfig `yaml:"classifiers"`

	Chains map[string]chainConfig `yaml:"chains"`
}

type authorizerConfig struct {
	Type string `yaml:"type"`

	Token string `yaml:"token"`

	Issuer   string `yaml:"issuer"`
	Audience string `yaml:"audience"`
}

type providerConfig struct {
	Type string `yaml:"type"`

	URL   string `yaml:"url"`
	Token string `yaml:"token"`

	Models map[string]modelConfig `yaml:"models"`
}

type modelConfig struct {
	ID string `yaml:"id"`

	System   string `yaml:"system"`
	Template string `yaml:"template"`

	Name        string `yaml:"name"`
	Description string `yaml:"description"`
}

type indexConfig struct {
	Type string `yaml:"type"`

	URL   string `yaml:"url"`
	Token string `yaml:"token"`

	Namespace string `yaml:"namespace"`
	Embedding string `yaml:"embedding"`
}

type extracterConfig struct {
	Type string `yaml:"type"`

	URL   string `yaml:"url"`
	Token string `yaml:"token"`
}

type classifierConfig struct {
	Type string `yaml:"type"`

	Model     string `yaml:"model"`
	Embedding string `yaml:"embedding"`

	Categories map[string]string `yaml:"categories"`
}

type chainConfig struct {
	Type string `yaml:"type"`

	Index string `yaml:"index"`

	Model     string `yaml:"model"`
	Embedding string `yaml:"embedding"`

	System   string `yaml:"system"`
	Template string `yaml:"template"`

	Limit    *int     `yaml:"limit"`
	Distance *float32 `yaml:"distance"`

	Filters map[string]filterConfig `yaml:"filters"`
}

type filterConfig struct {
	Classifier string `yaml:"classifier"`
}
