package orchestrator

import (
	"context"
	"fmt"

	"github.com/ThibLOREK/GoLoadlt/internal/etl/engine"
	"github.com/ThibLOREK/GoLoadlt/internal/jobs"
	"github.com/ThibLOREK/GoLoadlt/internal/xml/store"
)

// Service orchestre l'exécution des projets ETL.
// Il fait le lien entre l'API HTTP, le store XML et le moteur d'exécution DAG.
type Service struct {
	executor *engine.Executor
	xmlStore *store.XMLStore
	jobRepo  jobs.Repository
}

// NewService crée un nouveau Service d'orchestration.
func NewService(executor *engine.Executor, xmlStore *store.XMLStore, jobRepo jobs.Repository) *Service {
	return &Service{
		executor: executor,
		xmlStore: xmlStore,
		jobRepo:  jobRepo,
	}
}

// RunProject charge le projet XML, parse le DAG et l'exécute.
// Crée un Run en base, le passe à "running", exécute, puis persiste le statut final.
func (s *Service) RunProject(ctx context.Context, projectID string) (*engine.ExecutionReport, error) {
	// 1. Charger le projet depuis le store XML
	project, err := s.xmlStore.Load(projectID)
	if err != nil {
		return nil, fmt.Errorf("orchestrator: chargement projet %q: %w", projectID, err)
	}

	// 2. Créer l'entrée de run en base (status = pending)
	run, err := s.jobRepo.Create(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("orchestrator: création run: %w", err)
	}

	// 3. Passer en status "running"
	if setErr := s.jobRepo.SetStatus(ctx, run.ID, jobs.Running); setErr != nil {
		// Non-bloquant : on continue l'exécution même si la mise à jour échoue
		_ = setErr
	}

	// 4. Exécuter le DAG
	report, execErr := s.executor.Execute(ctx, project)

	// 5. Persister le statut final
	finalStatus := jobs.Succeeded
	if execErr != nil {
		finalStatus = jobs.Failed
	}
	if setErr := s.jobRepo.SetStatus(ctx, run.ID, finalStatus); setErr != nil {
		// Logguer mais ne pas masquer l'erreur d'exécution principale
		_ = setErr
	}

	if execErr != nil {
		return report, fmt.Errorf("orchestrator: exécution projet %q (run %s): %w", projectID, run.ID, execErr)
	}
	return report, nil
}

// CancelRun marque un run comme annulé.
func (s *Service) CancelRun(ctx context.Context, runID string) error {
	if err := s.jobRepo.SetStatus(ctx, runID, jobs.Cancelled); err != nil {
		return fmt.Errorf("orchestrator: annulation run %q: %w", runID, err)
	}
	return nil
}

// GetRunStatus retourne le run courant depuis la base.
func (s *Service) GetRunStatus(ctx context.Context, runID string) (*jobs.Run, error) {
	run, err := s.jobRepo.GetByID(ctx, runID)
	if err != nil {
		return nil, fmt.Errorf("orchestrator: récupération run %q: %w", runID, err)
	}
	return run, nil
}

// ListRuns retourne tous les runs d'un projet.
func (s *Service) ListRuns(ctx context.Context, projectID string) ([]jobs.Run, error) {
	runs, err := s.jobRepo.ListByProject(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("orchestrator: liste runs projet %q: %w", projectID, err)
	}
	return runs, nil
}
