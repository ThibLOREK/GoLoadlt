package orchestrator

import (
	"context"
	"fmt"

	"github.com/rinjold/go-etl-studio/internal/etl/engine"
	"github.com/rinjold/go-etl-studio/internal/jobs"
	"github.com/rinjold/go-etl-studio/internal/xml/store"
)

// Service orchestre l'exécution des projets ETL.
// Il fait le lien entre l'API HTTP, le store XML et le moteur d'exécution DAG.
type Service struct {
	executor  *engine.Executor
	xmlStore  *store.ProjectStore
	jobRepo   jobs.Repository
}

// NewService crée un nouveau Service d'orchestration.
// jobRepo peut être nil jusqu'à Sprint C (implémentation PostgreSQL).
func NewService(executor *engine.Executor, xmlStore *store.ProjectStore, jobRepo jobs.Repository) *Service {
	return &Service{
		executor: executor,
		xmlStore: xmlStore,
		jobRepo:  jobRepo,
	}
}

// RunProject charge le projet XML, parse le DAG et l'exécute.
// Si jobRepo est disponible : crée un Run en base, le passe à "running",
// exécute, puis persiste le statut final.
func (s *Service) RunProject(ctx context.Context, projectID string) (*engine.ExecutionReport, error) {
	// 1. Charger le projet depuis le store XML
	project, err := s.xmlStore.Load(projectID)
	if err != nil {
		return nil, fmt.Errorf("orchestrator: chargement projet %q: %w", projectID, err)
	}

	var runID string

	// 2. Créer l'entrée de run en base (si jobRepo disponible)
	if s.jobRepo != nil {
		run, createErr := s.jobRepo.Create(ctx, projectID)
		if createErr != nil {
			return nil, fmt.Errorf("orchestrator: création run: %w", createErr)
		}
		runID = run.ID
		_ = s.jobRepo.SetStatus(ctx, runID, jobs.Running)
	}

	// 3. Exécuter le DAG
	report, execErr := s.executor.Execute(ctx, project)

	// 4. Persister le statut final
	if s.jobRepo != nil && runID != "" {
		finalStatus := jobs.Succeeded
		if execErr != nil {
			finalStatus = jobs.Failed
		}
		_ = s.jobRepo.SetStatus(ctx, runID, finalStatus)
	}

	if execErr != nil {
		return report, fmt.Errorf("orchestrator: exécution projet %q: %w", projectID, execErr)
	}
	return report, nil
}

// CancelRun marque un run comme annulé.
func (s *Service) CancelRun(ctx context.Context, runID string) error {
	if s.jobRepo == nil {
		return fmt.Errorf("orchestrator: jobRepo non disponible (Sprint C)")
	}
	if err := s.jobRepo.SetStatus(ctx, runID, jobs.Cancelled); err != nil {
		return fmt.Errorf("orchestrator: annulation run %q: %w", runID, err)
	}
	return nil
}

// GetRunStatus retourne le run courant depuis la base.
func (s *Service) GetRunStatus(ctx context.Context, runID string) (*jobs.Run, error) {
	if s.jobRepo == nil {
		return nil, fmt.Errorf("orchestrator: jobRepo non disponible (Sprint C)")
	}
	run, err := s.jobRepo.GetByID(ctx, runID)
	if err != nil {
		return nil, fmt.Errorf("orchestrator: récupération run %q: %w", runID, err)
	}
	return run, nil
}

// ListRuns retourne tous les runs d'un projet.
func (s *Service) ListRuns(ctx context.Context, projectID string) ([]jobs.Run, error) {
	if s.jobRepo == nil {
		return []jobs.Run{}, nil
	}
	runs, err := s.jobRepo.ListByProject(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("orchestrator: liste runs projet %q: %w", projectID, err)
	}
	return runs, nil
}
