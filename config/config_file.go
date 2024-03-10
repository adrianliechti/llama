package config

import (
	"errors"
	"os"
	"strings"

	"github.com/adrianliechti/llama/pkg/prompt"
	"github.com/adrianliechti/llama/pkg/provider"
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
	Extractors  map[string]extractorConfig  `yaml:"extractors"`
	Classifiers map[string]classifierConfig `yaml:"classifiers"`

	Tools  map[string]toolConfig  `yaml:"tools"`
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

type ModelType string

const (
	ModelTypeCompleter   ModelType = "completer"
	ModelTypeEmbedder    ModelType = "embedder"
	ModelTypeTranscriber ModelType = "transcriber"
)

type modelConfig struct {
	ID   string    `yaml:"id"`
	Type ModelType `yaml:"type"`

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

type extractorConfig struct {
	Type string `yaml:"type"`

	URL   string `yaml:"url"`
	Token string `yaml:"token"`

	ChunkSize    *int `yaml:"chunkSize"`
	ChunkOverlap *int `yaml:"chunkOverlap"`
}

type classifierConfig struct {
	Type string `yaml:"type"`

	Model string `yaml:"model"`

	Template string    `yaml:"template"`
	Messages []message `yaml:"messages"`

	Classes map[string]string `yaml:"classes"`
}

type chainConfig struct {
	Type string `yaml:"type"`

	Model     string `yaml:"model"`
	Index     string `yaml:"index"`
	Embedding string `yaml:"embedding"`

	Template string    `yaml:"template"`
	Messages []message `yaml:"messages"`

	Tools []string `yaml:"tools"`

	Limit    *int     `yaml:"limit"`
	Distance *float32 `yaml:"distance"`

	Filters map[string]filterConfig `yaml:"filters"`
}

type toolConfig struct {
	Type string `yaml:"type"`

	URL   string `yaml:"url"`
	Token string `yaml:"token"`

	Model string `yaml:"model"`
	Index string `yaml:"index"`
}

type filterConfig struct {
	Classifier string `yaml:"classifier"`
}

type message struct {
	Role    string `yaml:"role"`
	Content string `yaml:"content"`
}

func parseTemplate(val string) (*prompt.Template, error) {
	if val == "" {
		return nil, errors.New("empty prompt")
	}

	if data, err := os.ReadFile(val); err == nil {
		return prompt.NewTemplate(string(data))
	}

	return prompt.NewTemplate(val)
}

func parseMessages(messages []message) ([]provider.Message, error) {
	result := make([]provider.Message, 0)

	for _, m := range messages {
		message, err := parseMessage(m)

		if err != nil {
			return nil, err

		}

		result = append(result, *message)
	}

	return result, nil
}

func parseMessage(message message) (*provider.Message, error) {
	var role provider.MessageRole

	if strings.EqualFold(message.Role, string(provider.MessageRoleSystem)) {
		role = provider.MessageRoleSystem
	}

	if strings.EqualFold(message.Role, string(provider.MessageRoleUser)) {
		role = provider.MessageRoleUser
	}

	if strings.EqualFold(message.Role, string(provider.MessageRoleAssistant)) {
		role = provider.MessageRoleAssistant
	}

	if role == "" {
		return nil, errors.New("invalid message role: " + message.Role)
	}

	return &provider.Message{
		Role:    role,
		Content: message.Content,
	}, nil
}
