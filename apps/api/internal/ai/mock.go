package ai

import (
	"context"

	"github.com/DontSpillTheTea/barrel/apps/api/internal/models"
)

type MockProvider struct{}

func NewMockProvider() *MockProvider {
	return &MockProvider{}
}

func (m *MockProvider) Name() string {
	return "mock_ai"
}

func (m *MockProvider) SecondRead(ctx context.Context, input SecondReadInput) (*models.AISecondRead, error) {
	// Dummy response for the mock provider
	return &models.AISecondRead{
		Eligible: true,
		Used:     true,
		Provider: "ai_native_mock",
		Candidates: models.AINativeExtraction{
			BrandName: models.AIExtractedField{
				Value:      "Mock Brand",
				Confidence: 0.99,
				Source:     "image",
			},
			ClassType: models.AIExtractedField{
				Value:      "Mock Type",
				Confidence: 0.98,
				Source:     "image",
			},
			AlcoholContent: models.AIAlcoholContent{
				ABV:        "42% Alc./Vol.",
				Proof:      "84 Proof",
				Confidence: 0.95,
				Source:     "image",
			},
			NetContents: models.AIExtractedField{
				Value:      "750 mL",
				Confidence: 0.96,
				Source:     "image",
			},
			GovernmentWarning: models.AIGovernmentWarning{
				Present:      true,
				VerbatimText: "GOVERNMENT WARNING: (1) According to the Surgeon General, women should not drink alcoholic beverages during pregnancy because of the risk of birth defects. (2) Consumption of alcoholic beverages impairs your ability to drive a car or operate machinery, and may cause health problems.",
				Confidence:   0.92,
				Source:       "image",
			},
			ProducerOrBottler: models.AIExtractedField{
				Value:      "Mock Distillery, Louisville, KY",
				Confidence: 0.90,
				Source:     "image",
			},
			CountryOfOrigin: models.AIExtractedField{
				Value:      "United States",
				Confidence: 0.95,
				Source:     "image",
			},
		},
	}, nil
}
