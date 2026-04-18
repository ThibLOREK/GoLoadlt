package app

import (
	"context"
	"time"

	"github.com/rinjold/go-etl-studio/internal/storage"
	"github.com/rinjold/go-etl-studio/pkg/models"
)

type WorkerApp struct {
	container *Container
}

func NewWorkerApp() (*WorkerApp, error) {
	container, err := BuildContainer(context.Background())
	if err != nil {
		return nil, err
	}
	return &WorkerApp{container: container}, nil
}

func (w *WorkerApp) Run(ctx context.Context) error {
	log := w.container.Logger.With().Str("component", "worker").Logger()
	log.Info().Msg("worker started")

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("worker stopping")
			return nil
		case <-ticker.C:
			w.processPendingRuns(ctx)
		}
	}
}

func (w *WorkerApp) processPendingRuns(ctx context.Context) {
	log := w.container.Logger.With().Str("component", "worker").Logger()

	pool := w.container.PostgresPool
	rows, err := pool.Query(ctx, `
		SELECT id, pipeline_id FROM runs WHERE status = 'pending'
		ORDER BY created_at ASC LIMIT 5
	`)
	if err != nil {
		log.Error().Err(err).Msg("failed to fetch pending runs")
		return
	}
	defer rows.Close()

	type job struct{ id, pipelineID string }
	jobs := make([]job, 0)
	for rows.Next() {
		var j job
		if err := rows.Scan(&j.id, &j.pipelineID); err == nil {
			jobs = append(jobs, j)
		}
	}

	runRepo := storage.NewRunRepository(pool)

	for _, j := range jobs {
		log.Info().Str("run_id", j.id).Msg("executing run")
		_ = runRepo.UpdateStatus(ctx, j.id, models.RunRunning, "")

		// Placeholder: real executor would be built from pipeline definition
		// For now we simulate a successful no-op run
		time.Sleep(200 * time.Millisecond)
		_ = runRepo.UpdateCounts(ctx, j.id, 0, 0)
		_ = runRepo.UpdateStatus(ctx, j.id, models.RunSucceeded, "")
		log.Info().Str("run_id", j.id).Msg("run succeeded")
	}
}
