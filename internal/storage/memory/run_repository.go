package memory

import (
	"context"
	"sync"
	"time"

	"github.com/rinjold/go-etl-studio/internal/storage"
	"github.com/rinjold/go-etl-studio/pkg/models"
)

type RunRepository struct {
	mu    sync.RWMutex
	store map[string]models.Run
}

func NewRunRepository() *RunRepository {
	return &RunRepository{store: make(map[string]models.Run)}
}

func (r *RunRepository) Create(_ context.Context, run models.Run) (models.Run, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	run.CreatedAt = time.Now().UTC()
	r.store[run.ID] = run
	return run, nil
}

func (r *RunRepository) GetByID(_ context.Context, id string) (models.Run, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	run, ok := r.store[id]
	if !ok {
		return models.Run{}, storage.ErrRunNotFound
	}
	return run, nil
}

func (r *RunRepository) ListByPipeline(_ context.Context, pipelineID string) ([]models.Run, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var runs []models.Run
	for _, run := range r.store {
		if run.PipelineID == pipelineID {
			runs = append(runs, run)
		}
	}
	return runs, nil
}

func (r *RunRepository) UpdateStatus(_ context.Context, id string, status models.RunStatus, errMsg string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	run, ok := r.store[id]
	if !ok {
		return storage.ErrRunNotFound
	}
	now := time.Now().UTC()
	run.Status = status
	switch status {
	case models.RunRunning:
		run.StartedAt = &now
	case models.RunSucceeded, models.RunFailed, models.RunCancelled:
		run.FinishedAt = &now
		run.ErrorMsg = errMsg
	}
	r.store[id] = run
	return nil
}

func (r *RunRepository) UpdateCounts(_ context.Context, id string, read, loaded int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	run, ok := r.store[id]
	if !ok {
		return storage.ErrRunNotFound
	}
	run.RecordsRead = read
	run.RecordsLoad = loaded
	r.store[id] = run
	return nil
}

func (r *RunRepository) ListPending(_ context.Context) ([]models.Run, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var runs []models.Run
	for _, run := range r.store {
		if run.Status == models.RunPending {
			runs = append(runs, run)
		}
	}
	return runs, nil
}
