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
	Auth authConfig `yaml:"auth"`

	Providers []providerConfig `yaml:"providers"`

	Chains map[string]chainConfig `yaml:"chains"`
}

type authConfig struct {
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

	Prompt   string `yaml:"prompt"`
	Template string `yaml:"template"`

	Name        string `yaml:"name"`
	Description string `yaml:"description"`
}

type chainConfig struct {
	Type string `yaml:"type"`

	Model     string `yaml:"model"`
	Embedding string `yaml:"embedding"`

	Index *indexConfig `yaml:"index"`
}

type indexConfig struct {
	Type string `yaml:"type"`

	URL   string `yaml:"url"`
	Token string `yaml:"token"`

	Name string `yaml:"name"`
}
