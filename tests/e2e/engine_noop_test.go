package e2e

import (
	"context"
	"testing"

	"github.com/rinjold/go-etl-studio/internal/etl/contracts"
	"github.com/rinjold/go-etl-studio/internal/etl/engine"
	"github.com/rinjold/go-etl-studio/internal/etl/transformers"
	"github.com/rs/zerolog"
)

type staticExtractor struct{ records []contracts.Record }
func (e staticExtractor) Extract(_ context.Context) ([]contracts.Record, error) { return e.records, nil }

type captureLoader struct{ captured []contracts.Record }
func (l *captureLoader) Load(_ context.Context, in []contracts.Record) error {
	l.captured = in
	return nil
}

func TestEngine_NoopTransformer(t *testing.T) {
	loader := &captureLoader{}
	ex := engine.Executor{
		Extractor:   staticExtractor{records: []contracts.Record{{"x": 1}, {"x": 2}}},
		Transformer: transformers.Noop{},
		Loader:      loader,
		Log:         zerolog.Nop(),
	}

	result := ex.Execute(context.Background())
	if result.Err != nil {
		t.Fatal(result.Err)
	}
	if len(loader.captured) != 2 {
		t.Errorf("expected 2 records, got %d", len(loader.captured))
	}
}
