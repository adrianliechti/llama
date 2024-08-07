package config

import (
	"errors"
	"sort"
	"strings"

	"github.com/adrianliechti/llama/pkg/provider"
)

func (cfg *Config) RegisterModel(id string) {
	if cfg.models == nil {
		cfg.models = make(map[string]provider.Model)
	}

	if _, ok := cfg.models[id]; ok {
		return
	}

	cfg.models[id] = provider.Model{
		ID: id,
	}
}

func (cfg *Config) Models() []provider.Model {
	var result []provider.Model

	for _, m := range cfg.models {
		result = append(result, m)
	}

	sort.SliceStable(result, func(i, j int) bool { return result[i].ID < result[j].ID })

	return result
}

func (cfg *Config) Model(id string) (*provider.Model, error) {
	if cfg.models != nil {
		if m, ok := cfg.models[id]; ok {
			return &m, nil
		}
	}

	return nil, errors.New("model not found: " + id)
}

type ModelType string

const (
	ModelTypeAuto        ModelType = ""
	ModelTypeCompleter   ModelType = "completer"
	ModelTypeEmbedder    ModelType = "embedder"
	ModelTypeRenderer    ModelType = "renderer"
	ModelTypeSynthesizer ModelType = "synthesizer"
	ModelTypeTranscriber ModelType = "transcriber"
	ModelTypeTranslator  ModelType = "translator"
)

type modelConfig struct {
	ID string `yaml:"id"`

	Type ModelType `yaml:"type"`

	Name        string `yaml:"name"`
	Description string `yaml:"description"`
}

type modelContext struct {
	ID string

	Type ModelType

	Name        string
	Description string
}

func detectModelType(id string) ModelType {
	completers := []string{
		"aya",
		"claude",
		"codestral",
		"command",
		"deepseek",
		"dolphin",
		"falcon",
		"gemini",
		"gemma",
		"gpt",
		"hermes",
		"llama",
		"llava",
		"mistral",
		"mixtral",
		"orca",
		"phi",
		"qwen",
		"stable-code",
		"stablelm",
		"starcoder",
		"vicuna",
		"wizardlm",
		"zephyr",
	}

	embedders := []string{
		"bge",
		"embed",
		"gte",
		"minilm",
	}

	renderers := []string{
		"dall-e",
		"sd-turbo",
		"sdxl",
		"stable-diffusion",
	}

	synthesizers := []string{
		"stable-audio",
		"tts",
	}

	transcribers := []string{
		"whisper",
	}

	for _, val := range completers {
		if strings.Contains(strings.ToLower(id), strings.ToLower(val)) {
			return ModelTypeCompleter
		}
	}

	for _, val := range embedders {
		if strings.Contains(strings.ToLower(id), strings.ToLower(val)) {
			return ModelTypeEmbedder
		}
	}

	for _, val := range renderers {
		if strings.Contains(strings.ToLower(id), strings.ToLower(val)) {
			return ModelTypeRenderer
		}
	}

	for _, val := range synthesizers {
		if strings.Contains(strings.ToLower(id), strings.ToLower(val)) {
			return ModelTypeSynthesizer
		}
	}

	for _, val := range transcribers {
		if strings.Contains(strings.ToLower(id), strings.ToLower(val)) {
			return ModelTypeTranscriber
		}
	}

	return ModelTypeAuto
}
