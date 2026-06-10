package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
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
	baseDir := "data/reviews"
	os.MkdirAll(baseDir, 0755)
	return &LocalProvider{baseDir: baseDir}
}

func (l *LocalProvider) getJobDir(jobID string) string {
	dir := filepath.Join(l.baseDir, jobID)
	os.MkdirAll(dir, 0755)
	return dir
}

func (l *LocalProvider) SaveImage(ctx context.Context, jobID string, data []byte) error {
	return os.WriteFile(filepath.Join(l.getJobDir(jobID), "image.png"), data, 0644)
}

func (l *LocalProvider) SaveResult(ctx context.Context, jobID string, result *models.LabelAnalysisResult) error {
	b, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(l.getJobDir(jobID), "result.json"), b, 0644)
}

func (l *LocalProvider) SaveDecision(ctx context.Context, jobID string, decision ReviewDecision) error {
	b, err := json.MarshalIndent(decision, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(l.getJobDir(jobID), "decision.json"), b, 0644)
}

func (l *LocalProvider) ListReviews(ctx context.Context) ([]ReviewRecord, error) {
	entries, err := os.ReadDir(l.baseDir)
	if err != nil {
		return nil, err
	}

	var records []ReviewRecord
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		jobID := entry.Name()
		record, err := l.GetReview(ctx, jobID)
		if err == nil && record != nil {
			records = append(records, *record)
		}
	}

	// Sort newest first
	sort.Slice(records, func(i, j int) bool {
		return records[i].Timestamp > records[j].Timestamp
	})

	return records, nil
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
