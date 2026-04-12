package files

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/aeshield/backend/models"
)

type JobRepo interface {
	Create(ctx context.Context, job *models.Job) error
	FindByID(ctx context.Context, id string) (*models.Job, error)
	FindByUserID(ctx context.Context, userID string, status string, limit, offset int64) ([]*models.Job, int64, error)
	Update(ctx context.Context, job *models.Job) error
	MarkCancelled(ctx context.Context, id string) error
	DeleteOlderThan(ctx context.Context, before time.Time) error
}

var ErrJobCancelled = errors.New("job cancelled")

const JobExecutionTimeout = 15 * time.Minute

type JobService struct {
	repo JobRepo
	mu   sync.Mutex
}

func NewJobService(repo JobRepo) *JobService {
	return &JobService{repo: repo}
}

func (s *JobService) CreateJob(ctx context.Context, userID string, jobType models.JobType, filename string) (*models.Job, error) {
	job := &models.Job{
		UserID:   userID,
		Type:     jobType,
		Status:   models.JobStatusPending,
		Progress: 0,
		Filename: filename,
	}
	if err := s.repo.Create(ctx, job); err != nil {
		return nil, err
	}
	return job, nil
}

func (s *JobService) Start(ctx context.Context, jobID string) error {
	job, err := s.repo.FindByID(ctx, jobID)
	if err != nil {
		return err
	}
	job.Status = models.JobStatusProcessing
	return s.repo.Update(ctx, job)
}

func (s *JobService) UpdateProgress(ctx context.Context, jobID string, progress int) error {
	if progress < 0 {
		progress = 0
	}
	if progress > 100 {
		progress = 100
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	job, err := s.repo.FindByID(ctx, jobID)
	if err != nil {
		return err
	}
	if job.Status == models.JobStatusCancelled || job.Status == models.JobStatusCompleted || job.Status == models.JobStatusFailed {
		return nil
	}
	job.Status = models.JobStatusProcessing
	job.Progress = progress
	return s.repo.Update(ctx, job)
}

func (s *JobService) Complete(ctx context.Context, jobID string, result *models.JobResult, fileMetadataID *string) error {
	job, err := s.repo.FindByID(ctx, jobID)
	if err != nil {
		return err
	}
	now := time.Now()
	job.Status = models.JobStatusCompleted
	job.Progress = 100
	job.Result = result
	job.CompletedAt = &now
	return s.repo.Update(ctx, job)
}

func (s *JobService) Fail(ctx context.Context, jobID, code, message string) error {
	job, err := s.repo.FindByID(ctx, jobID)
	if err != nil {
		return err
	}
	now := time.Now()
	job.Status = models.JobStatusFailed
	job.Error = &models.JobError{Code: code, Message: message, Timestamp: now}
	job.CompletedAt = &now
	return s.repo.Update(ctx, job)
}

func (s *JobService) Cancel(ctx context.Context, jobID string) error {
	return s.repo.MarkCancelled(ctx, jobID)
}

func (s *JobService) GetJob(ctx context.Context, jobID string) (*models.Job, error) {
	return s.repo.FindByID(ctx, jobID)
}

func (s *JobService) IsCancelled(ctx context.Context, jobID string) (bool, error) {
	job, err := s.repo.FindByID(ctx, jobID)
	if err != nil {
		return false, err
	}
	return job.Status == models.JobStatusCancelled, nil
}

func (s *JobService) ListJobs(ctx context.Context, userID string, status string, limit, offset int64) ([]*models.Job, int64, error) {
	return s.repo.FindByUserID(ctx, userID, status, limit, offset)
}

func (s *JobService) CleanupOldJobs(ctx context.Context, olderThan time.Duration) error {
	return s.repo.DeleteOlderThan(ctx, time.Now().Add(-olderThan))
}

func progressFromBytes(processed, total int64) int {
	if total <= 0 {
		return 0
	}
	if processed <= 0 {
		return 0
	}
	if processed >= total {
		return 100
	}
	return int((processed * 100) / total)
}

func ensurePasswordForJob(password string) error {
	if password == "" {
		return fmt.Errorf("password is required")
	}
	return nil
}
