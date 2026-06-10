package storage

import (
	"context"
	"os"

	"github.com/DontSpillTheTea/barrel/apps/api/internal/models"
)

type ReviewDecision struct {
	Decision string `json:"decision"`
	Notes    string `json:"notes"`
}

type ReviewRecord struct {
	JobID        string                     `json:"job_id"`
	Filename     string                     `json:"filename"`
	Timestamp    string                     `json:"timestamp"`
	Status       string                     `json:"status"` // unreviewed, approved, rejected, needs_more_info
	Notes        string                     `json:"notes"`
	Result       *models.LabelAnalysisResult `json:"result,omitempty"`
	HasImage     bool                       `json:"has_image"`
}

type Provider interface {
	SaveImage(ctx context.Context, jobID string, data []byte) error
	SaveResult(ctx context.Context, jobID string, result *models.LabelAnalysisResult) error
	SaveDecision(ctx context.Context, jobID string, decision ReviewDecision) error
	ListReviews(ctx context.Context) ([]ReviewRecord, error)
	GetReview(ctx context.Context, jobID string) (*ReviewRecord, error)
	GetImage(ctx context.Context, jobID string) ([]byte, error)
}

func NewProvider() Provider {
	providerType := os.Getenv("STORAGE_PROVIDER")
	if providerType == "azure" {
		return NewAzureBlobProvider()
	}
	return NewLocalProvider()
}
