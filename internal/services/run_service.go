package services

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/rinjold/go-etl-studio/pkg/models"
)

type RunRepository interface {
	Create(ctx context.Context, run models.Run) (models.Run, error)
	GetByID(ctx context.Context, id string) (models.Run, error)
	ListByPipeline(ctx context.Context, pipelineID string) ([]models.Run, error)
	ListPending(ctx context.Context) ([]models.Run, error)
	UpdateStatus(ctx context.Context, id string, status models.RunStatus, errMsg string) error
	UpdateCounts(ctx context.Context, id string, read, loaded int64) error
}

type RunService struct {
	repo     RunRepository
	pipeRepo PipelineRepository
}

func NewRunService(runRepo RunRepository, pipeRepo PipelineRepository) *RunService {
	return &RunService{repo: runRepo, pipeRepo: pipeRepo}
}

func (s *RunService) Schedule(ctx context.Context, pipelineID string) (models.Run, error) {
	if _, err := s.pipeRepo.GetByID(ctx, pipelineID); err != nil {
		return models.Run{}, err
	}
	run := models.Run{
		ID:         uuid.NewString(),
		PipelineID: pipelineID,
		Status:     models.RunPending,
		CreatedAt:  time.Now().UTC(),
	}
	return s.repo.Create(ctx, run)
}

func (s *RunService) GetByID(ctx context.Context, id string) (models.Run, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *RunService) ListByPipeline(ctx context.Context, pipelineID string) ([]models.Run, error) {
	return s.repo.ListByPipeline(ctx, pipelineID)
}

func (s *RunService) ListPending(ctx context.Context) ([]models.Run, error) {
	return s.repo.ListPending(ctx)
}

func (s *RunService) UpdateStatus(ctx context.Context, id string, status models.RunStatus, errMsg string) error {
	return s.repo.UpdateStatus(ctx, id, status, errMsg)
}

func (s *RunService) UpdateCounts(ctx context.Context, id string, read, loaded int64) error {
	return s.repo.UpdateCounts(ctx, id, read, loaded)
}
