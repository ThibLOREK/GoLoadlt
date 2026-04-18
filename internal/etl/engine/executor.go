package engine

import (
	"context"
	"fmt"
	"time"

	"github.com/rinjold/go-etl-studio/internal/etl/contracts"
	"github.com/rs/zerolog"
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
	start := time.Now()

	e.Log.Info().Msg("extraction started")
	records, err := e.Extractor.Extract(ctx)
	if err != nil {
		return RunResult{Err: fmt.Errorf("extract: %w", err), Duration: time.Since(start)}
	}
	e.Log.Info().Int("records", len(records)).Msg("extraction done")

	if e.Transformer != nil {
		records, err = e.Transformer.Transform(ctx, records)
		if err != nil {
			return RunResult{
				RecordsRead: int64(len(records)),
				Err:         fmt.Errorf("transform: %w", err),
				Duration:    time.Since(start),
			}
		}
	}

	e.Log.Info().Msg("loading started")
	if err := e.Loader.Load(ctx, records); err != nil {
		return RunResult{
			RecordsRead: int64(len(records)),
			Err:         fmt.Errorf("load: %w", err),
			Duration:    time.Since(start),
		}
	}

	return RunResult{
		RecordsRead:   int64(len(records)),
		RecordsLoaded: int64(len(records)),
		Duration:      time.Since(start),
	}
}
