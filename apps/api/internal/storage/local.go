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

func (l *LocalProvider) ListReviews(ctx context.Context) ([]ReviewRecord, error) {
	log.Printf("Listing reviews from: %s", l.baseDir)
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
