package jobs

import (
	"testing"
	"time"

	"github.com/DontSpillTheTea/barrel/apps/api/internal/models"
)

func TestJobStore(t *testing.T) {
	store := NewStore()

	// Test create
	job := store.CreateJob("test.png")
	if job.ID == "" {
		t.Errorf("Expected job ID to be set")
	}
	if job.Status != StatusQueued {
		t.Errorf("Expected status to be queued, got %s", job.Status)
	}
	if job.Filename != "test.png" {
		t.Errorf("Expected filename to be test.png, got %s", job.Filename)
	}

	// Test get
	fetched, ok := store.GetJob(job.ID)
	if !ok {
		t.Errorf("Expected to find job by ID")
	}
	if fetched.ID != job.ID {
		t.Errorf("Expected fetched job ID to match")
	}

	// Test update status
	store.UpdateJobStatus(job.ID, StatusProcessing)
	fetched, _ = store.GetJob(job.ID)
	if fetched.Status != StatusProcessing {
		t.Errorf("Expected status to be updated to processing")
	}

	// Test succeed
	res := &models.LabelAnalysisResult{
		OverallStatus: "Pass",
	}
	store.SucceedJob(job.ID, res)
	fetched, _ = store.GetJob(job.ID)
	if fetched.Status != StatusSucceeded {
		t.Errorf("Expected status to be succeeded")
	}
	if fetched.Result.OverallStatus != "Pass" {
		t.Errorf("Expected result to be set")
	}

	// Test fail
	job2 := store.CreateJob("test2.png")
	store.FailJob(job2.ID, "some error", nil)
	fetched2, _ := store.GetJob(job2.ID)
	if fetched2.Status != StatusFailed {
		t.Errorf("Expected status to be failed")
	}
	if fetched2.Error != "some error" {
		t.Errorf("Expected error to be set")
	}

	// Test concurrency (simple)
	go store.CreateJob("concurrent1.png")
	go store.CreateJob("concurrent2.png")
	time.Sleep(10 * time.Millisecond) // Give goroutines time to run
	
	store.RLock()
	if len(store.jobs) != 4 {
		t.Errorf("Expected 4 total jobs, got %d", len(store.jobs))
	}
	store.RUnlock()
}
