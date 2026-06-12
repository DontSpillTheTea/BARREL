package config

import (
	"os"
	"strconv"
)

type Config struct {
	OCRWorkerURL          string
	MaxUploadMB           int64
	RulesPath             string
	AINativeEnabled       bool
	AzureOpenAIEndpoint   string
	AzureOpenAIAPIKey     string
	AzureOpenAIDeployment string
	AzureOpenAIAPIVersion string

	AnalysisProvider        string
	TextParserDeployment    string
	EscalationEnabled       bool
	OcrMinConfidence        float64
	FieldMinConfidence      float64
	GovWarningMinSimilarity float64

	AzureStorageTable string
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

func envFloat(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if f, err := strconv.ParseFloat(value, 64); err == nil {
			return f
		}
	}
	return defaultValue
}

func envString(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
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

		AnalysisProvider:        envString("BARREL_ANALYSIS_PROVIDER", "tiered"),
		TextParserDeployment:    envString("BARREL_TEXT_PARSER_DEPLOYMENT", ""),
		EscalationEnabled:       envBool("BARREL_ESCALATION_ENABLED", "", true),
		OcrMinConfidence:        envFloat("BARREL_OCR_MIN_CONFIDENCE", 0.80),
		FieldMinConfidence:      envFloat("BARREL_FIELD_MIN_CONFIDENCE", 0.70),
		GovWarningMinSimilarity: envFloat("BARREL_GOV_WARNING_MIN_SIMILARITY", 0.95),

		AzureStorageTable: envString("AZURE_STORAGE_TABLE", "reviews"),
	}
}
