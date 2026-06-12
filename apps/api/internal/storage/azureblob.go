package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/data/aztables"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/DontSpillTheTea/barrel/apps/api/internal/models"
)

type AzureBlobProvider struct {
	blobClient  *azblob.Client
	tableClient *aztables.Client
	container   string
	tableName   string
}

func NewAzureBlobProvider() *AzureBlobProvider {
	connStr := os.Getenv("AZURE_STORAGE_CONNECTION_STRING")
	container := os.Getenv("AZURE_STORAGE_CONTAINER")
	if container == "" {
		container = "jobs"
	}
	tableName := os.Getenv("AZURE_STORAGE_TABLE")
	if tableName == "" {
		tableName = "reviews"
	}

	p := &AzureBlobProvider{container: container, tableName: tableName}

	if connStr == "" {
		log.Println("WARNING: AZURE_STORAGE_CONNECTION_STRING not set")
		return p
	}

	blobClient, err := azblob.NewClientFromConnectionString(connStr, nil)
	if err != nil {
		log.Printf("WARNING: Failed to create blob client: %v", err)
		return p
	}
	p.blobClient = blobClient

	// Auto-create blob container if it doesn't exist (needed for floci-az / fresh environments)
	ctx := context.Background()
	if _, err := blobClient.CreateContainer(ctx, container, nil); err != nil {
		log.Printf("Blob container '%s' already exists or create skipped: %v", container, err)
	} else {
		log.Printf("Created blob container: %s", container)
	}

	serviceClient, err := aztables.NewServiceClientFromConnectionString(connStr, nil)
	if err != nil {
		log.Printf("WARNING: Failed to create table client: %v", err)
	} else {
		p.tableClient = serviceClient.NewClient(tableName)

		// Auto-create table if it doesn't exist
		if _, err := serviceClient.CreateTable(ctx, tableName, nil); err != nil {
			log.Printf("Table '%s' already exists or create skipped: %v", tableName, err)
		} else {
			log.Printf("Created table: %s", tableName)
		}

		log.Printf("Azure Table Storage configured (table: %s)", tableName)
	}

	log.Printf("Azure Blob provider configured (container: %s)", container)
	return p
}

func (a *AzureBlobProvider) upload(ctx context.Context, blobName string, data []byte) error {
	if a.blobClient == nil {
		return fmt.Errorf("blob client not configured")
	}
	_, err := a.blobClient.UploadBuffer(ctx, a.container, blobName, data, nil)
	return err
}

func (a *AzureBlobProvider) download(ctx context.Context, blobName string) ([]byte, error) {
	if a.blobClient == nil {
		return nil, fmt.Errorf("blob client not configured")
	}
	resp, err := a.blobClient.DownloadStream(ctx, a.container, blobName, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

func (a *AzureBlobProvider) upsertMetadata(ctx context.Context, summary models.ReviewSummary) error {
	if a.tableClient == nil {
		return nil
	}

	submitted, _ := time.Parse(time.RFC3339, summary.SubmittedAt)
	partitionKey := submitted.Format("2006-01")
	if partitionKey == "0001-01" {
		partitionKey = time.Now().Format("2006-01")
	}

	entity := map[string]interface{}{
		"PartitionKey":      partitionKey,
		"RowKey":            summary.JobID,
		"JobID":             summary.JobID,
		"Filename":          summary.Filename,
		"SubmittedAt":       summary.SubmittedAt,
		"CompletedAt":       summary.CompletedAt,
		"ProviderRequested": summary.ProviderRequested,
		"ProviderUsed":      summary.ProviderUsed,
		"OverallStatus":     summary.OverallStatus,
		"OverallConfidence": summary.OverallConfidence,
		"FieldPassCount":    summary.FieldPassCount,
		"FieldTotalCount":   summary.FieldTotalCount,
		"ReviewerDecision":  summary.ReviewerDecision,
		"BeverageType":      summary.BeverageType,
		"BrandName":         summary.BrandName,
		"ClassType":         summary.ClassType,
		"AlcoholContent":    summary.AlcoholContent,
		"NetContents":       summary.NetContents,
		"HasImage":          true,
	}

	entityBytes, err := json.Marshal(entity)
	if err != nil {
		return err
	}

	_, err = a.tableClient.UpsertEntity(ctx, entityBytes, &aztables.UpsertEntityOptions{
		UpdateMode: aztables.UpdateModeMerge,
	})
	return err
}

func (a *AzureBlobProvider) SaveImage(ctx context.Context, jobID string, data []byte) error {
	log.Printf("Azure Blob: saving image %s/image.png", jobID)
	return a.upload(ctx, jobID+"/image.png", data)
}

func (a *AzureBlobProvider) SaveResult(ctx context.Context, jobID string, result *models.LabelAnalysisResult) error {
	b, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}
	log.Printf("Azure Blob: saving result %s/result.json", jobID)
	if err := a.upload(ctx, jobID+"/result.json", b); err != nil {
		return err
	}

	summary := resultToSummary(jobID, result)
	if err := a.upsertMetadata(ctx, summary); err != nil {
		log.Printf("Azure Table: failed to upsert metadata for %s: %v", jobID, err)
	}
	return nil
}

func (a *AzureBlobProvider) SaveDecision(ctx context.Context, jobID string, decision ReviewDecision) error {
	b, err := json.MarshalIndent(decision, "", "  ")
	if err != nil {
		return err
	}
	log.Printf("Azure Blob: saving decision %s/decision.json", jobID)
	if err := a.upload(ctx, jobID+"/decision.json", b); err != nil {
		return err
	}

	if a.tableClient != nil {
		pager := a.tableClient.NewListEntitiesPager(&aztables.ListEntitiesOptions{
			Filter: to_ptr(fmt.Sprintf("RowKey eq '%s'", jobID)),
			Top:    to_ptr(int32(1)),
		})
		for pager.More() {
			page, err := pager.NextPage(ctx)
			if err != nil {
				break
			}
			for _, raw := range page.Entities {
				var entity map[string]interface{}
				if json.Unmarshal(raw, &entity) == nil {
					entity["ReviewerDecision"] = decision.Decision
					updated, _ := json.Marshal(entity)
					a.tableClient.UpsertEntity(ctx, updated, &aztables.UpsertEntityOptions{
						UpdateMode: aztables.UpdateModeMerge,
					})
				}
			}
		}
	}
	return nil
}

func (a *AzureBlobProvider) ListReviews(ctx context.Context) ([]models.ReviewSummary, error) {
	if a.tableClient == nil {
		return nil, fmt.Errorf("table client not configured")
	}

	var summaries []models.ReviewSummary
	pager := a.tableClient.NewListEntitiesPager(nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		for _, raw := range page.Entities {
			var entity map[string]interface{}
			if err := json.Unmarshal(raw, &entity); err != nil {
				continue
			}
			summaries = append(summaries, entityToSummary(entity))
		}
	}

	return summaries, nil
}

func (a *AzureBlobProvider) GetReview(ctx context.Context, jobID string) (*ReviewRecord, error) {
	resultData, err := a.download(ctx, jobID+"/result.json")
	if err != nil {
		return nil, fmt.Errorf("review not found: %v", err)
	}

	var res models.LabelAnalysisResult
	if err := json.Unmarshal(resultData, &res); err != nil {
		return nil, fmt.Errorf("corrupt result.json: %v", err)
	}

	record := &ReviewRecord{
		JobID:     jobID,
		Filename:  res.Filename,
		Status:    "unreviewed",
		Timestamp: time.Now().Format(time.RFC3339),
		Result:    &res,
		HasImage:  true,
	}

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

func resultToSummary(jobID string, result *models.LabelAnalysisResult) models.ReviewSummary {
	s := models.ReviewSummary{
		ID:                jobID,
		JobID:             jobID,
		Filename:          result.Filename,
		SubmittedAt:       time.Now().Format(time.RFC3339),
		ProviderRequested: result.RequestedProvider,
		OverallStatus:     result.OverallStatus,
		OverallConfidence: result.OverallConfidence,
		BeverageType:      result.BeverageType,
		BrandName:         result.ExtractedFields.BrandName,
		ClassType:         result.ExtractedFields.ClassType,
		AlcoholContent:    result.ExtractedFields.AlcoholContent,
		NetContents:       result.ExtractedFields.NetContents,
	}

	providerUsed := "unknown"
	if result.AIEscalation.Used {
		providerUsed = result.AIEscalation.Provider
	} else if result.OCR != nil {
		providerUsed = result.OCR.SelectedProvider
	} else if result.RequestedProvider != "" {
		providerUsed = result.RequestedProvider
	}
	s.ProviderUsed = providerUsed

	passCount := 0
	for _, f := range result.Fields {
		if f.Status == models.StatusMatch || f.Status == "Pass" {
			passCount++
		}
	}
	s.FieldPassCount = passCount
	s.FieldTotalCount = len(result.Fields)

	return s
}

func entityToSummary(entity map[string]interface{}) models.ReviewSummary {
	s := models.ReviewSummary{}
	if v, ok := entity["JobID"].(string); ok {
		s.ID = v
		s.JobID = v
	}
	if v, ok := entity["Filename"].(string); ok {
		s.Filename = v
	}
	if v, ok := entity["SubmittedAt"].(string); ok {
		s.SubmittedAt = v
	}
	if v, ok := entity["CompletedAt"].(string); ok {
		s.CompletedAt = v
	}
	if v, ok := entity["ProviderRequested"].(string); ok {
		s.ProviderRequested = v
	}
	if v, ok := entity["ProviderUsed"].(string); ok {
		s.ProviderUsed = v
	}
	if v, ok := entity["OverallStatus"].(string); ok {
		s.OverallStatus = v
	}
	if v, ok := entity["OverallConfidence"]; ok {
		s.OverallConfidence = toInt(v)
	}
	if v, ok := entity["FieldPassCount"]; ok {
		s.FieldPassCount = toInt(v)
	}
	if v, ok := entity["FieldTotalCount"]; ok {
		s.FieldTotalCount = toInt(v)
	}
	if v, ok := entity["ReviewerDecision"].(string); ok {
		s.ReviewerDecision = v
	}
	if v, ok := entity["BeverageType"].(string); ok {
		s.BeverageType = v
	}
	if v, ok := entity["BrandName"].(string); ok {
		s.BrandName = v
	}
	if v, ok := entity["ClassType"].(string); ok {
		s.ClassType = v
	}
	if v, ok := entity["AlcoholContent"].(string); ok {
		s.AlcoholContent = v
	}
	if v, ok := entity["NetContents"].(string); ok {
		s.NetContents = v
	}
	return s
}

func toInt(v interface{}) int {
	switch val := v.(type) {
	case float64:
		return int(val)
	case int:
		return val
	case string:
		n, _ := strconv.Atoi(val)
		return n
	default:
		return 0
	}
}

func to_ptr[T any](v T) *T { return &v }

var _ Provider = (*AzureBlobProvider)(nil)
