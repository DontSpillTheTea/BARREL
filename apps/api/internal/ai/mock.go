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
		Provider: "mock_ai",
		Findings: []models.AISecondReadFinding{
			{Field: "general", Severity: "info", Message: "The label appears to comply with formatting requirements based on AI review."},
			{Field: "general", Severity: "info", Message: "Mock AI successfully read the label context."},
		},
		Candidates: models.AISecondReadCandidates{
			BrandName: "Mock Brand",
			ClassType: "Mock Type",
		},
	}, nil
}
