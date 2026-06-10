package config

import "os"

type Config struct {
	OCRWorkerURL               string
	MaxUploadMB                int64
	RulesPath                  string
	AISecondReadEnabled        bool
	AISecondReadAutoOnFail     bool
	AzureOpenAIEndpoint        string
	AzureOpenAIAPIKey          string
	AzureOpenAIDeployment      string
	AzureOpenAIAPIVersion      string
	AISecondReadTimeoutSeconds int
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
		OCRWorkerURL:               url,
		MaxUploadMB:                25,
		RulesPath:                  rulesPath,
		AISecondReadEnabled:        os.Getenv("AI_SECOND_READ_ENABLED") == "true",
		AISecondReadAutoOnFail:     os.Getenv("AI_SECOND_READ_AUTO_ON_FAIL") == "true",
		AzureOpenAIEndpoint:        os.Getenv("AZURE_OPENAI_ENDPOINT"),
		AzureOpenAIAPIKey:          os.Getenv("AZURE_OPENAI_API_KEY"),
		AzureOpenAIDeployment:      os.Getenv("AZURE_OPENAI_DEPLOYMENT"),
		AzureOpenAIAPIVersion:      os.Getenv("AZURE_OPENAI_API_VERSION"),
		AISecondReadTimeoutSeconds: 45,
	}
}
