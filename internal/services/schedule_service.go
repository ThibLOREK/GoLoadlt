package services

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/rinjold/go-etl-studio/internal/etl/scheduler"
	"github.com/rinjold/go-etl-studio/pkg/models"
)

type ScheduleRepository interface {
	Upsert(ctx context.Context, s models.Schedule) (models.Schedule, error)
	GetByPipeline(ctx context.Context, pipelineID string) (models.Schedule, error)
	ListDue(ctx context.Context, now time.Time) ([]models.Schedule, error)
	UpdateAfterFire(ctx context.Context, id string, lastRun time.Time, nextRun *time.Time) error
	Delete(ctx context.Context, pipelineID string) error
}

type ScheduleService struct {
	repo    ScheduleRepository
	runSvc  *RunService
}

func NewScheduleService(repo ScheduleRepository, runSvc *RunService) *ScheduleService {
	return &ScheduleService{repo: repo, runSvc: runSvc}
}

func (s *ScheduleService) Upsert(ctx context.Context, pipelineID, cronExpr string, enabled bool) (models.Schedule, error) {
	next, err := scheduler.Next(cronExpr, time.Now())
	if err != nil {
		return models.Schedule{}, err
	}

	existing, err := s.repo.GetByPipeline(ctx, pipelineID)
	id := uuid.NewString()
	if err == nil {
		id = existing.ID
	}

	return s.repo.Upsert(ctx, models.Schedule{
		ID:         id,
		PipelineID: pipelineID,
		CronExpr:   cronExpr,
		Enabled:    enabled,
		NextRunAt:  &next,
	})
}

func (s *ScheduleService) GetByPipeline(ctx context.Context, pipelineID string) (models.Schedule, error) {
	return s.repo.GetByPipeline(ctx, pipelineID)
}

func (s *ScheduleService) Delete(ctx context.Context, pipelineID string) error {
	return s.repo.Delete(ctx, pipelineID)
}

func (s *ScheduleService) Tick(ctx context.Context) error {
	now := time.Now().UTC()
	due, err := s.repo.ListDue(ctx, now)
	if err != nil {
		return err
	}

	for _, sched := range due {
		if _, err := s.runSvc.Schedule(ctx, sched.PipelineID); err != nil {
			continue
		}

		next, err := scheduler.Next(sched.CronExpr, now)
		var nextPtr *time.Time
		if err == nil {
			nextPtr = &next
		}

		_ = s.repo.UpdateAfterFire(ctx, sched.ID, now, nextPtr)
	}

	return nil
}
