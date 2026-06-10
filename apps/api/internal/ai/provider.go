package ai

import (
	"context"

	"github.com/DontSpillTheTea/barrel/apps/api/internal/models"
)

type SecondReadInput struct {
	Filename       string
	ContentType    string
	ImageBytes     []byte
	OCRText        string
	ExpectedFields models.ExpectedLabelFields
	BeverageType   string
	InitialResult  models.LabelAnalysisResult
}

type Provider interface {
	Name() string
	SecondRead(ctx context.Context, input SecondReadInput) (*models.AISecondRead, error)
}
