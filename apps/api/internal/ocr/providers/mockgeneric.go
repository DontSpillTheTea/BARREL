package providers

import (
	"context"
	"fmt"
)

type MockGenericProvider struct{}

func NewMockGenericProvider() *MockGenericProvider {
	return &MockGenericProvider{}
}

func (m *MockGenericProvider) Name() string {
	return "mock_generic"
}

func (m *MockGenericProvider) Extract(ctx context.Context, input ExtractInput) (*OCRResult, error) {
	mockText := fmt.Sprintf(`BLACK MAPLE RESERVE
Kentucky Straight Bourbon Whiskey
Distilled and Bottled by Black Maple Distillery, Louisville, Kentucky
45%% Alc./Vol. (90 Proof)
750 mL
Product of United States

GOVERNMENT WARNING: (1) According to the Surgeon General, women should not drink alcoholic beverages during pregnancy because of the risk of birth defects. (2) Consumption of alcoholic beverages impairs your ability to drive a car or operate machinery, and may cause health problems.

File: %s`, input.Filename)

	return &OCRResult{
		Status:           "ok",
		Provider:         "mock_generic",
		SelectedProvider: "mock_generic",
		Text:             mockText,
		MeanConfidence:   85.0,
	}, nil
}
