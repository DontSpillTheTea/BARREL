package config

import "os"

type Config struct {
	OCRWorkerURL          string
	MaxUploadMB           int64
	RulesPath             string
	AINativeEnabled       bool
	AzureOpenAIEndpoint   string
	AzureOpenAIAPIKey     string
	AzureOpenAIDeployment string
	AzureOpenAIAPIVersion string
}

func envBool(primary, legacy string, defaultValue bool) bool {
	if value := os.Getenv(primary); value != "" {
		return value == "true"
	}
	if legacy != "" {
		if value := os.Getenv(legacy); value != "" {
			return value == "true"
		}
	}
	return defaultValue
}

func Load() Config {
	url := os.Getenv("OCR_WORKER_URL")
	if url == "" {
		url = "http://localhost:9090"
	}
	rulesPath := os.Getenv("RULESET_PATH")
	if rulesPath == "" {
		rulesPath = "../../rules/ttb"
	}
	return Config{
		OCRWorkerURL:          url,
		MaxUploadMB:           25,
		RulesPath:             rulesPath,
		AINativeEnabled:       envBool("AI_NATIVE_ENABLED", "AI_SECOND_READ_ENABLED", true),
		AzureOpenAIEndpoint:   os.Getenv("AZURE_OPENAI_ENDPOINT"),
		AzureOpenAIAPIKey:     os.Getenv("AZURE_OPENAI_API_KEY"),
		AzureOpenAIDeployment: os.Getenv("AZURE_OPENAI_DEPLOYMENT"),
		AzureOpenAIAPIVersion: os.Getenv("AZURE_OPENAI_API_VERSION"),
	}
}
