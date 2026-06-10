package providers

import (
	"context"
)

type ExtractInput struct {
	Filename    string
	ContentType string
	Data        []byte
}

type OCRResult struct {
	Status           string         `json:"status"`
	Provider         string         `json:"provider"`
	SelectedProvider string         `json:"selected_provider"`
	Text             string         `json:"text"`
	MeanConfidence   float64        `json:"mean_confidence"`
	Raw              map[string]any `json:"raw"`
	Warnings         []string       `json:"warnings"`
}

type Provider interface {
	Name() string
	Extract(ctx context.Context, input ExtractInput) (*OCRResult, error)
}
