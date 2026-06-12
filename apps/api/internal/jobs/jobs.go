package jobs

import (
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/DontSpillTheTea/barrel/apps/api/internal/models"
)

type Status string

const (
	StatusQueued     Status = "queued"
	StatusProcessing Status = "processing"
	StatusSucceeded  Status = "succeeded"
	StatusFailed     Status = "failed"
)

type Job struct {
	ID        string                      `json:"job_id"`
	Status    Status                      `json:"status"`
	CreatedAt time.Time                   `json:"created_at"`
	UpdatedAt time.Time                   `json:"updated_at"`
	Filename  string                      `json:"filename"`
	Error     string                      `json:"error,omitempty"`
	Result    *models.LabelAnalysisResult `json:"result,omitempty"`
	Decision  string                      `json:"decision,omitempty"`
	Notes     string                      `json:"notes,omitempty"`
}

type Store struct {
	sync.RWMutex
	jobs map[string]*Job
}

func NewStore() *Store {
	return &Store{
		jobs: make(map[string]*Job),
	}
}

func (s *Store) CreateJob(filename string) *Job {
	s.Lock()
	defer s.Unlock()

	id := uuid.New().String()
	now := time.Now()

	job := &Job{
		ID:        id,
		Status:    StatusQueued,
		CreatedAt: now,
		UpdatedAt: now,
		Filename:  filename,
	}

	s.jobs[id] = job
	return job
}

func (s *Store) GetJob(id string) (*Job, bool) {
	s.RLock()
	defer s.RUnlock()

	job, ok := s.jobs[id]
	return job, ok
}

func (s *Store) UpdateJobStatus(id string, status Status) {
	s.Lock()
	defer s.Unlock()

	if job, ok := s.jobs[id]; ok {
		job.Status = status
		job.UpdatedAt = time.Now()
	}
}

func (s *Store) FailJob(id string, errMsg string, result *models.LabelAnalysisResult) {
	s.Lock()
	defer s.Unlock()

	if job, ok := s.jobs[id]; ok {
		job.Status = StatusFailed
		job.Error = errMsg
		job.Result = result
		job.UpdatedAt = time.Now()
	}
}

func (s *Store) SucceedJob(id string, result *models.LabelAnalysisResult) {
	s.Lock()
	defer s.Unlock()

	if job, ok := s.jobs[id]; ok {
		job.Status = StatusSucceeded
		job.Result = result
		job.UpdatedAt = time.Now()
	}
}

func (s *Store) SetDecision(id, decision, notes string) bool {
	s.Lock()
	defer s.Unlock()

	if job, ok := s.jobs[id]; ok {
		job.Decision = decision
		job.Notes = notes
		job.UpdatedAt = time.Now()
		return true
	}
	return false
}

func (s *Store) ListJobs() []*Job {
	s.RLock()
	defer s.RUnlock()

	out := make([]*Job, 0, len(s.jobs))
	for _, job := range s.jobs {
		out = append(out, job)
	}
	return out
}
