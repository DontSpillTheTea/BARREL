package providers

import (
	"context"
	"errors"
)

type MockFastProvider struct{}

func NewMockFastProvider() *MockFastProvider {
	return &MockFastProvider{}
}

func (m *MockFastProvider) Name() string {
	return "mock_fast"
}

func (m *MockFastProvider) Extract(ctx context.Context, input ExtractInput) (*OCRResult, error) {
	if input.Filename == "good_07_brand_case_variation.png" {
		return &OCRResult{
			Status:           "ok",
			Provider:         "mock_fast",
			SelectedProvider: "mock_fast",
			Text:             "STONE'S THROW SPIRITS\nRye Whiskey\n46% Alc./Vol. (92 Proof)\n750 mL\nGOVERNMENT WARNING: (1) According to the Surgeon General, women should not drink alcoholic beverages during pregnancy because of the risk of birth defects. (2) Consumption of alcoholic beverages impairs your ability to drive a car or operate machinery, and may cause health problems.",
			MeanConfidence:   99.9,
			Raw:              map[string]any{"mock": true},
			Warnings:         nil,
		}, nil
	}
	return nil, errors.New("mock not available for this image")
}
