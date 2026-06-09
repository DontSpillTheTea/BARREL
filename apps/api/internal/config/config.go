package config

import "os"

type Config struct {
	OCRWorkerURL string
	MaxUploadMB  int64
	RulesPath    string
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
		OCRWorkerURL: url,
		MaxUploadMB:  25,
		RulesPath:    rulesPath,
	}
}
