package services

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/rinjold/go-etl-studio/pkg/models"
)

type PipelineRepository interface {
	List(ctx context.Context) ([]models.Pipeline, error)
	GetByID(ctx context.Context, id string) (models.Pipeline, error)
	Create(ctx context.Context, p models.Pipeline) (models.Pipeline, error)
	Update(ctx context.Context, p models.Pipeline) (models.Pipeline, error)
	Delete(ctx context.Context, id string) error
}

type PipelineService struct {
	repo PipelineRepository
}

func NewPipelineService(repo PipelineRepository) *PipelineService {
	return &PipelineService{repo: repo}
}

func (s *PipelineService) List(ctx context.Context) ([]models.Pipeline, error) {
	return s.repo.List(ctx)
}

func (s *PipelineService) GetByID(ctx context.Context, id string) (models.Pipeline, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *PipelineService) Create(ctx context.Context, input CreatePipelineInput) (models.Pipeline, error) {
	now := time.Now().UTC()
	pipeline := models.Pipeline{
		ID:          uuid.NewString(),
		Name:        input.Name,
		Description: input.Description,
		Status:      "draft",
		SourceType:  input.SourceType,
		TargetType:  input.TargetType,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	return s.repo.Create(ctx, pipeline)
}

func (s *PipelineService) Update(ctx context.Context, id string, input UpdatePipelineInput) (models.Pipeline, error) {
	current, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return models.Pipeline{}, err
	}

	current.Name = input.Name
	current.Description = input.Description
	current.Status = input.Status
	current.SourceType = input.SourceType
	current.TargetType = input.TargetType
	current.UpdatedAt = time.Now().UTC()

	return s.repo.Update(ctx, current)
}

func (s *PipelineService) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

type CreatePipelineInput struct {
	Name        string
	Description string
	SourceType  string
	TargetType  string
}

type UpdatePipelineInput struct {
	Name        string
	Description string
	Status      string
	SourceType  string
	TargetType  string
}
