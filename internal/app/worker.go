package app

import (
	"context"
	"encoding/json"
	"time"

	"github.com/rinjold/go-etl-studio/internal/etl/engine"
	etlpipeline "github.com/rinjold/go-etl-studio/internal/etl/pipeline"
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

	runs, err := w.container.RunService.ListPending(ctx)
	if err != nil {
		log.Error().Err(err).Msg("failed to fetch pending runs")
		return
	}

	for _, run := range runs {
		runLog := log.With().Str("run_id", run.ID).Str("pipeline_id", run.PipelineID).Logger()

		_ = w.container.RunService.UpdateStatus(ctx, run.ID, models.RunRunning, "")
		telemetry.ActiveRuns.Inc()

		pipe, err := w.container.PipelineService.GetByID(ctx, run.PipelineID)
		if err != nil {
			runLog.Error().Err(err).Msg("pipeline not found")
			_ = w.container.RunService.UpdateStatus(ctx, run.ID, models.RunFailed, err.Error())
			telemetry.ActiveRuns.Dec()
			telemetry.RecordRun(run.PipelineID, "failed", 0, 0, 0)
			continue
		}

		def, err := pipelineToDefinition(pipe)
		if err != nil {
			runLog.Error().Err(err).Msg("invalid pipeline definition")
			_ = w.container.RunService.UpdateStatus(ctx, run.ID, models.RunFailed, err.Error())
			telemetry.ActiveRuns.Dec()
			telemetry.RecordRun(run.PipelineID, "failed", 0, 0, 0)
			continue
		}

		executor, err := engine.BuildExecutor(ctx, def, w.container.PostgresPool, runLog)
		if err != nil {
			runLog.Error().Err(err).Msg("failed to build executor")
			_ = w.container.RunService.UpdateStatus(ctx, run.ID, models.RunFailed, err.Error())
			telemetry.ActiveRuns.Dec()
			telemetry.RecordRun(run.PipelineID, "failed", 0, 0, 0)
			continue
		}

		runLog.Info().Msg("executing pipeline")
		result := executor.Execute(ctx)

		telemetry.ActiveRuns.Dec()
		_ = w.container.RunService.UpdateCounts(ctx, run.ID, result.RecordsRead, result.RecordsLoaded)

		if result.Err != nil {
			runLog.Error().Err(result.Err).Dur("duration", result.Duration).Msg("run failed")
			_ = w.container.RunService.UpdateStatus(ctx, run.ID, models.RunFailed, result.Err.Error())
			telemetry.RecordRun(run.PipelineID, "failed", result.Duration, result.RecordsRead, result.RecordsLoaded)
		} else {
			runLog.Info().
				Int64("read", result.RecordsRead).
				Int64("loaded", result.RecordsLoaded).
				Dur("duration", result.Duration).
				Msg("run succeeded")
			_ = w.container.RunService.UpdateStatus(ctx, run.ID, models.RunSucceeded, "")
			telemetry.RecordRun(run.PipelineID, "succeeded", result.Duration, result.RecordsRead, result.RecordsLoaded)
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
