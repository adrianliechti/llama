package azure

import (
	"errors"
	"os"
	"regexp"
	"strings"

	"github.com/adrianliechti/llama/pkg/llm/openai"

	client "github.com/sashabaranov/go-openai"
)

func FromEnvironment() (*openai.Provider, error) {
	url := os.Getenv("AZURE_OPENAI_API_URL")

	if url == "" {
		return nil, errors.New("AZURE_OPENAI_API_URL is not configured")
	}

	token := os.Getenv("AZURE_OPENAI_API_KEY")

	if token == "" {
		return nil, errors.New("AZURE_OPENAI_API_KEY is not configured")
	}

	cfg := client.DefaultAzureConfig(token, url)

	deployments := getEnvMap("AZURE_OPENAI_DEPLOYMENTS")

	cfg.AzureModelMapperFunc = func(model string) string {
		if val, ok := deployments[model]; ok {
			return val
		}

		return regexp.MustCompile(`[.:]`).ReplaceAllString(model, "")
	}

	return openai.New(cfg)
}

func getEnvMap(key string) map[string]string {
	result := make(map[string]string)

	pairs := strings.Split(os.Getenv(key), ",")

	for _, val := range pairs {
		parts := strings.SplitN(val, "=", 2)

		if len(parts) == 2 {
			result[parts[0]] = parts[1]
		}
	}

	return result
}
