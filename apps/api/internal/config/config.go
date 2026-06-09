package config

import "os"

type Config struct {
	OCRWorkerURL string
	MaxUploadMB  int64
}

func Load() Config {
	url := os.Getenv("OCR_WORKER_URL")
	if url == "" {
		url = "http://localhost:9090"
	}
	return Config{
		OCRWorkerURL: url,
		MaxUploadMB:  25,
	}
}
