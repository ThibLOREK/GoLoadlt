package app

import (
	"context"
	"encoding/json"
	"time"

	"github.com/rinjold/go-etl-studio/internal/etl/engine"
	etlpipeline "github.com/rinjold/go-etl-studio/internal/etl/pipeline"
	"github.com/rinjold/go-etl-studio/internal/storage"
	"github.com/rinjold/go-etl-studio/internal/telemetry"
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

	jobTicker := time.NewTicker(5 * time.Second)
	cronTicker := time.NewTicker(30 * time.Second)
	defer jobTicker.Stop()
	defer cronTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("worker stopping")
			return nil
		case <-cronTicker.C:
			if err := w.container.ScheduleService.Tick(ctx); err != nil {
				log.Error().Err(err).Msg("scheduler tick failed")
			}
		case <-jobTicker.C:
			w.processPendingRuns(ctx)
		}
	}
}

func (w *WorkerApp) processPendingRuns(ctx context.Context) {
	log := w.container.Logger.With().Str("component", "worker").Logger()
	pool := w.container.PostgresPool
	runRepo := storage.NewRunRepository(pool)
	pipelineRepo := storage.NewPipelineRepository(pool)

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
	var jobs []job
	for rows.Next() {
		var j job
		if err := rows.Scan(&j.id, &j.pipelineID); err == nil {
			jobs = append(jobs, j)
		}
	}

	for _, j := range jobs {
		runLog := log.With().Str("run_id", j.id).Str("pipeline_id", j.pipelineID).Logger()
		_ = runRepo.UpdateStatus(ctx, j.id, models.RunRunning, "")
		telemetry.ActiveRuns.Inc()

		pipe, err := pipelineRepo.GetByID(ctx, j.pipelineID)
		if err != nil {
			runLog.Error().Err(err).Msg("pipeline not found")
			_ = runRepo.UpdateStatus(ctx, j.id, models.RunFailed, err.Error())
			telemetry.ActiveRuns.Dec()
			telemetry.RecordRun(j.pipelineID, "failed", 0, 0, 0)
			continue
		}

		def, err := pipelineToDefinition(pipe)
		if err != nil {
			runLog.Error().Err(err).Msg("invalid pipeline definition")
			_ = runRepo.UpdateStatus(ctx, j.id, models.RunFailed, err.Error())
			telemetry.ActiveRuns.Dec()
			telemetry.RecordRun(j.pipelineID, "failed", 0, 0, 0)
			continue
		}

		executor, err := engine.BuildExecutor(ctx, def, pool, runLog)
		if err != nil {
			runLog.Error().Err(err).Msg("failed to build executor")
			_ = runRepo.UpdateStatus(ctx, j.id, models.RunFailed, err.Error())
			telemetry.ActiveRuns.Dec()
			telemetry.RecordRun(j.pipelineID, "failed", 0, 0, 0)
			continue
		}

		runLog.Info().Msg("executing pipeline")
		result := executor.Execute(ctx)

		telemetry.ActiveRuns.Dec()
		_ = runRepo.UpdateCounts(ctx, j.id, result.RecordsRead, result.RecordsLoaded)

		if result.Err != nil {
			runLog.Error().Err(result.Err).Dur("duration", result.Duration).Msg("run failed")
			_ = runRepo.UpdateStatus(ctx, j.id, models.RunFailed, result.Err.Error())
			telemetry.RecordRun(j.pipelineID, "failed", result.Duration, result.RecordsRead, result.RecordsLoaded)
		} else {
			runLog.Info().
				Int64("read", result.RecordsRead).
				Int64("loaded", result.RecordsLoaded).
				Dur("duration", result.Duration).
				Msg("run succeeded")
			_ = runRepo.UpdateStatus(ctx, j.id, models.RunSucceeded, "")
			telemetry.RecordRun(j.pipelineID, "succeeded", result.Duration, result.RecordsRead, result.RecordsLoaded)
		}
	}
}

func pipelineToDefinition(p models.Pipeline) (etlpipeline.Definition, error) {
	var steps []etlpipeline.TransformStep
	if p.Steps != nil {
		if err := json.Unmarshal(p.Steps, &steps); err != nil {
			return etlpipeline.Definition{}, err
		}
	}
	return etlpipeline.Definition{
		ID:           p.ID,
		Name:         p.Name,
		SourceType:   etlpipeline.SourceType(p.SourceType),
		TargetType:   etlpipeline.TargetType(p.TargetType),
		SourceConfig: p.SourceConfig,
		TargetConfig: p.TargetConfig,
		Steps:        steps,
	}, nil
}
