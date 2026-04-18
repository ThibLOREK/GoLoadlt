package memory

import (
	"context"
	"sync"
	"time"

	"github.com/rinjold/go-etl-studio/internal/storage"
	"github.com/rinjold/go-etl-studio/pkg/models"
)

type ScheduleRepository struct {
	mu    sync.RWMutex
	store map[string]models.Schedule // keyed by pipelineID
}

func NewScheduleRepository() *ScheduleRepository {
	return &ScheduleRepository{store: make(map[string]models.Schedule)}
}

func (r *ScheduleRepository) Upsert(_ context.Context, s models.Schedule) (models.Schedule, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	now := time.Now().UTC()
	if existing, ok := r.store[s.PipelineID]; ok {
		s.CreatedAt = existing.CreatedAt
	} else {
		s.CreatedAt = now
	}
	s.UpdatedAt = now
	r.store[s.PipelineID] = s
	return s, nil
}

func (r *ScheduleRepository) GetByPipeline(_ context.Context, pipelineID string) (models.Schedule, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	s, ok := r.store[pipelineID]
	if !ok {
		return models.Schedule{}, storage.ErrScheduleNotFound
	}
	return s, nil
}

func (r *ScheduleRepository) ListDue(_ context.Context, now time.Time) ([]models.Schedule, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var due []models.Schedule
	for _, s := range r.store {
		if s.Enabled && s.NextRunAt != nil && !s.NextRunAt.After(now) {
			due = append(due, s)
		}
	}
	return due, nil
}

func (r *ScheduleRepository) UpdateAfterFire(_ context.Context, id string, lastRun time.Time, nextRun *time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for pid, s := range r.store {
		if s.ID == id {
			s.LastRunAt = &lastRun
			s.NextRunAt = nextRun
			s.UpdatedAt = time.Now().UTC()
			r.store[pid] = s
			return nil
		}
	}
	return storage.ErrScheduleNotFound
}

func (r *ScheduleRepository) Delete(_ context.Context, pipelineID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.store, pipelineID)
	return nil
}
