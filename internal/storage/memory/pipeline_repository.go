package memory

import (
	"context"
	"sync"
	"time"

	"github.com/rinjold/go-etl-studio/internal/storage"
	"github.com/rinjold/go-etl-studio/pkg/models"
)

type PipelineRepository struct {
	mu    sync.RWMutex
	store map[string]models.Pipeline
}

func NewPipelineRepository() *PipelineRepository {
	return &PipelineRepository{store: make(map[string]models.Pipeline)}
}

func (r *PipelineRepository) List(_ context.Context) ([]models.Pipeline, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]models.Pipeline, 0, len(r.store))
	for _, p := range r.store {
		list = append(list, p)
	}
	return list, nil
}

func (r *PipelineRepository) GetByID(_ context.Context, id string) (models.Pipeline, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.store[id]
	if !ok {
		return models.Pipeline{}, storage.ErrPipelineNotFound
	}
	return p, nil
}

func (r *PipelineRepository) Create(_ context.Context, p models.Pipeline) (models.Pipeline, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	now := time.Now().UTC()
	p.CreatedAt = now
	p.UpdatedAt = now
	r.store[p.ID] = p
	return p, nil
}

func (r *PipelineRepository) Update(_ context.Context, p models.Pipeline) (models.Pipeline, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.store[p.ID]; !ok {
		return models.Pipeline{}, storage.ErrPipelineNotFound
	}
	p.UpdatedAt = time.Now().UTC()
	r.store[p.ID] = p
	return p, nil
}

func (r *PipelineRepository) Delete(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.store[id]; !ok {
		return storage.ErrPipelineNotFound
	}
	delete(r.store, id)
	return nil
}
