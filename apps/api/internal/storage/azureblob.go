package storage

import (
	"context"
	"errors"

	"github.com/DontSpillTheTea/barrel/apps/api/internal/models"
)

type AzureBlobProvider struct{}

func NewAzureBlobProvider() *AzureBlobProvider {
	return &AzureBlobProvider{}
}

func (a *AzureBlobProvider) SaveImage(ctx context.Context, jobID string, data []byte) error {
	return errors.New("not implemented")
}

func (a *AzureBlobProvider) SaveResult(ctx context.Context, jobID string, result *models.LabelAnalysisResult) error {
	return errors.New("not implemented")
}

func (a *AzureBlobProvider) SaveDecision(ctx context.Context, jobID string, decision ReviewDecision) error {
	return errors.New("not implemented")
}

func (a *AzureBlobProvider) ListReviews(ctx context.Context) ([]ReviewRecord, error) {
	return nil, errors.New("not implemented")
}

func (a *AzureBlobProvider) GetReview(ctx context.Context, jobID string) (*ReviewRecord, error) {
	return nil, errors.New("not implemented")
}

func (a *AzureBlobProvider) GetImage(ctx context.Context, jobID string) ([]byte, error) {
	return nil, errors.New("not implemented")
}
