package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/DontSpillTheTea/barrel/apps/api/internal/models"
)

type LocalProvider struct {
	baseDir string
}

func NewLocalProvider() *LocalProvider {
	baseDir := os.Getenv("BARREL_STORAGE_DIR")
	if baseDir == "" {
		baseDir = "./data/barrel_storage"
	}
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		baseDir = filepath.Join(os.TempDir(), "barrel_storage")
		_ = os.MkdirAll(baseDir, 0755)
	}
	log.Printf("Local review storage base directory: %s", baseDir)
	return &LocalProvider{baseDir: baseDir}
}

func (l *LocalProvider) getJobDir(jobID string) (string, error) {
	dir := filepath.Join(l.baseDir, jobID)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	return dir, nil
}

func (l *LocalProvider) SaveImage(ctx context.Context, jobID string, data []byte) error {
	dir, err := l.getJobDir(jobID)
	if err != nil {
		return err
	}
	path := filepath.Join(dir, "image.png")
	log.Printf("Saving review image: %s", path)
	return os.WriteFile(path, data, 0644)
}

func (l *LocalProvider) SaveResult(ctx context.Context, jobID string, result *models.LabelAnalysisResult) error {
	b, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}
	dir, err := l.getJobDir(jobID)
	if err != nil {
		return err
	}
	path := filepath.Join(dir, "result.json")
	log.Printf("Saving review result: %s", path)
	return os.WriteFile(path, b, 0644)
}

func (l *LocalProvider) SaveDecision(ctx context.Context, jobID string, decision ReviewDecision) error {
	b, err := json.MarshalIndent(decision, "", "  ")
	if err != nil {
		return err
	}
	dir, err := l.getJobDir(jobID)
	if err != nil {
		return err
	}
	path := filepath.Join(dir, "decision.json")
	log.Printf("Saving review decision: %s", path)
	return os.WriteFile(path, b, 0644)
}

func (l *LocalProvider) ListReviews(ctx context.Context) ([]models.ReviewSummary, error) {
	log.Printf("Listing reviews from: %s", l.baseDir)
	entries, err := os.ReadDir(l.baseDir)
	if err != nil {
		return nil, err
	}

	var summaries []models.ReviewSummary
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		jobID := entry.Name()
		record, err := l.GetReview(ctx, jobID)
		if err == nil && record != nil {
			summaries = append(summaries, recordToLocalSummary(record))
		}
	}

	sort.Slice(summaries, func(i, j int) bool {
		return summaries[i].SubmittedAt > summaries[j].SubmittedAt
	})

	return summaries, nil
}

func recordToLocalSummary(r *ReviewRecord) models.ReviewSummary {
	s := models.ReviewSummary{
		ID:               r.JobID,
		JobID:            r.JobID,
		Filename:         r.Filename,
		SubmittedAt:      r.Timestamp,
		ReviewerDecision: r.Status,
	}
	if r.Result != nil {
		if s.Filename == "" {
			s.Filename = r.Result.Filename
		}
		s.OverallStatus = r.Result.OverallStatus
		s.OverallConfidence = r.Result.OverallConfidence
		s.BeverageType = r.Result.BeverageType
		s.ProviderRequested = r.Result.RequestedProvider
		s.BrandName = r.Result.ExtractedFields.BrandName
		s.ClassType = r.Result.ExtractedFields.ClassType
		s.AlcoholContent = r.Result.ExtractedFields.AlcoholContent
		s.NetContents = r.Result.ExtractedFields.NetContents

		providerUsed := "unknown"
		if r.Result.AIEscalation.Used {
			providerUsed = r.Result.AIEscalation.Provider
		} else if r.Result.OCR != nil {
			providerUsed = r.Result.OCR.SelectedProvider
		} else if r.Result.RequestedProvider != "" {
			providerUsed = r.Result.RequestedProvider
		}
		s.ProviderUsed = providerUsed

		passCount := 0
		for _, f := range r.Result.Fields {
			if f.Status == models.StatusMatch || f.Status == "Pass" {
				passCount++
			}
		}
		s.FieldPassCount = passCount
		s.FieldTotalCount = len(r.Result.Fields)
	}
	return s
}

func (l *LocalProvider) GetReview(ctx context.Context, jobID string) (*ReviewRecord, error) {
	dir := filepath.Join(l.baseDir, jobID)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil, fmt.Errorf("review not found")
	}

	record := &ReviewRecord{
		JobID:  jobID,
		Status: "unreviewed",
	}

	info, err := os.Stat(dir)
	if err == nil {
		record.Timestamp = info.ModTime().Format(time.RFC3339)
	}

	if _, err := os.Stat(filepath.Join(dir, "image.png")); err == nil {
		record.HasImage = true
	}

	if b, err := os.ReadFile(filepath.Join(dir, "result.json")); err == nil {
		var res models.LabelAnalysisResult
		if err := json.Unmarshal(b, &res); err == nil {
			record.Result = &res
			record.Filename = res.Filename
		}
	}

	if b, err := os.ReadFile(filepath.Join(dir, "decision.json")); err == nil {
		var dec ReviewDecision
		if err := json.Unmarshal(b, &dec); err == nil {
			record.Status = dec.Decision
			record.Notes = dec.Notes
		}
	}

	return record, nil
}

func (l *LocalProvider) GetImage(ctx context.Context, jobID string) ([]byte, error) {
	file, err := os.Open(filepath.Join(l.baseDir, jobID, "image.png"))
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return io.ReadAll(file)
}
