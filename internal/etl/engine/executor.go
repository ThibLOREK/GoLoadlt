package engine

import (
	"context"
	"fmt"
	"time"

	"github.com/rinjold/go-etl-studio/internal/etl/contracts"
	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

type RunResult struct {
	RecordsRead   int64
	RecordsLoaded int64
	Duration      time.Duration
	Err           error
}

type Executor struct {
	Extractor   contracts.Extractor
	Transformer contracts.Transformer
	Loader      contracts.Loader
	Log         zerolog.Logger
}

func (e Executor) Execute(ctx context.Context) RunResult {
	tracer := otel.Tracer("etl.engine")
	ctx, span := tracer.Start(ctx, "pipeline.execute")
	defer span.End()

	start := time.Now()

	// Extract
	_, extractSpan := tracer.Start(ctx, "extract")
	e.Log.Info().Msg("extraction started")
	records, err := e.Extractor.Extract(ctx)
	extractSpan.End()
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return RunResult{Err: fmt.Errorf("extract: %w", err), Duration: time.Since(start)}
	}
	e.Log.Info().Int("records", len(records)).Msg("extraction done")

	// Transform
	if e.Transformer != nil {
		_, transformSpan := tracer.Start(ctx, "transform")
		records, err = e.Transformer.Transform(ctx, records)
		transformSpan.End()
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return RunResult{RecordsRead: int64(len(records)), Err: fmt.Errorf("transform: %w", err), Duration: time.Since(start)}
		}
	}

	// Load
	_, loadSpan := tracer.Start(ctx, "load")
	e.Log.Info().Msg("loading started")
	err = e.Loader.Load(ctx, records)
	loadSpan.End()
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return RunResult{RecordsRead: int64(len(records)), Err: fmt.Errorf("load: %w", err), Duration: time.Since(start)}
	}

	span.SetAttributes(
		attribute.Int64("records.read", int64(len(records))),
		attribute.Int64("records.loaded", int64(len(records))),
	)
	span.SetStatus(codes.Ok, "")

	return RunResult{
		RecordsRead:   int64(len(records)),
		RecordsLoaded: int64(len(records)),
		Duration:      time.Since(start),
	}
}
