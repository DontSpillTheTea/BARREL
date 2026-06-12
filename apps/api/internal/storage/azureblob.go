package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/DontSpillTheTea/barrel/apps/api/internal/models"
)

type AzureBlobProvider struct {
	client        *azblob.Client
	containerName string
}

func NewAzureBlobProvider() *AzureBlobProvider {
	connStr := os.Getenv("AZURE_STORAGE_CONNECTION_STRING")
	container := os.Getenv("AZURE_STORAGE_CONTAINER")
	if container == "" {
		container = "jobs"
	}

	if connStr == "" {
		log.Println("WARNING: AZURE_STORAGE_CONNECTION_STRING not set, Azure Blob provider will fail")
		return &AzureBlobProvider{containerName: container}
	}

	client, err := azblob.NewClientFromConnectionString(connStr, nil)
	if err != nil {
		log.Printf("WARNING: Failed to create Azure Blob client: %v", err)
		return &AzureBlobProvider{containerName: container}
	}

	log.Printf("Azure Blob provider configured (container: %s)", container)
	return &AzureBlobProvider{client: client, containerName: container}
}

func (a *AzureBlobProvider) upload(ctx context.Context, blobName string, data []byte, contentType string) error {
	if a.client == nil {
		return fmt.Errorf("azure blob client not configured")
	}
	_, err := a.client.UploadBuffer(ctx, a.containerName, blobName, data, &azblob.UploadBufferOptions{})
	return err
}

func (a *AzureBlobProvider) download(ctx context.Context, blobName string) ([]byte, error) {
	if a.client == nil {
		return nil, fmt.Errorf("azure blob client not configured")
	}
	resp, err := a.client.DownloadStream(ctx, a.containerName, blobName, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

func (a *AzureBlobProvider) SaveImage(ctx context.Context, jobID string, data []byte) error {
	blobName := jobID + "/image.png"
	log.Printf("Azure Blob: saving image %s", blobName)
	return a.upload(ctx, blobName, data, "image/png")
}

func (a *AzureBlobProvider) SaveResult(ctx context.Context, jobID string, result *models.LabelAnalysisResult) error {
	b, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}
	blobName := jobID + "/result.json"
	log.Printf("Azure Blob: saving result %s", blobName)
	return a.upload(ctx, blobName, b, "application/json")
}

func (a *AzureBlobProvider) SaveDecision(ctx context.Context, jobID string, decision ReviewDecision) error {
	b, err := json.MarshalIndent(decision, "", "  ")
	if err != nil {
		return err
	}
	blobName := jobID + "/decision.json"
	log.Printf("Azure Blob: saving decision %s", blobName)
	return a.upload(ctx, blobName, b, "application/json")
}

func (a *AzureBlobProvider) ListReviews(ctx context.Context) ([]ReviewRecord, error) {
	if a.client == nil {
		return nil, fmt.Errorf("azure blob client not configured")
	}

	jobIDs := map[string]bool{}
	pager := a.client.NewListBlobsFlatPager(a.containerName, &azblob.ListBlobsFlatOptions{})
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		for _, blob := range page.Segment.BlobItems {
			name := *blob.Name
			parts := strings.SplitN(name, "/", 2)
			if len(parts) == 2 {
				jobIDs[parts[0]] = true
			}
		}
	}

	var records []ReviewRecord
	for jobID := range jobIDs {
		record, err := a.GetReview(ctx, jobID)
		if err == nil && record != nil {
			records = append(records, *record)
		}
	}

	return records, nil
}

func (a *AzureBlobProvider) GetReview(ctx context.Context, jobID string) (*ReviewRecord, error) {
	record := &ReviewRecord{
		JobID:     jobID,
		Status:    "unreviewed",
		Timestamp: time.Now().Format(time.RFC3339),
	}

	resultData, err := a.download(ctx, jobID+"/result.json")
	if err != nil {
		return nil, fmt.Errorf("review not found: %v", err)
	}

	var res models.LabelAnalysisResult
	if err := json.Unmarshal(resultData, &res); err != nil {
		return nil, fmt.Errorf("corrupt result.json: %v", err)
	}
	record.Result = &res
	record.Filename = res.Filename

	// Check for image
	_, imgErr := a.download(ctx, jobID+"/image.png")
	record.HasImage = imgErr == nil

	// Check for decision
	decData, decErr := a.download(ctx, jobID+"/decision.json")
	if decErr == nil {
		var dec ReviewDecision
		if json.Unmarshal(decData, &dec) == nil {
			record.Status = dec.Decision
			record.Notes = dec.Notes
		}
	}

	return record, nil
}

func (a *AzureBlobProvider) GetImage(ctx context.Context, jobID string) ([]byte, error) {
	return a.download(ctx, jobID+"/image.png")
}

// Ensure AzureBlobProvider implements Provider
var _ Provider = (*AzureBlobProvider)(nil)
