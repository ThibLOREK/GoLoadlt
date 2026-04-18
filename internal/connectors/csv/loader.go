package csv

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"

	"github.com/rinjold/go-etl-studio/internal/etl/contracts"
)

type LoaderConfig struct {
	FilePath  string
	Delimiter rune
	HasHeader bool
	Columns   []string
}

type Loader struct {
	cfg LoaderConfig
}

func NewLoader(cfg LoaderConfig) *Loader {
	if cfg.Delimiter == 0 {
		cfg.Delimiter = ','
	}
	return &Loader{cfg: cfg}
}

func (l *Loader) Load(ctx context.Context, records []contracts.Record) error {
	if len(records) == 0 {
		return nil
	}

	f, err := os.Create(l.cfg.FilePath)
	if err != nil {
		return fmt.Errorf("csv loader: create file: %w", err)
	}
	defer f.Close()

	w := csv.NewWriter(f)
	w.Comma = l.cfg.Delimiter

	cols := l.cfg.Columns
	if len(cols) == 0 {
		for k := range records[0] {
			cols = append(cols, k)
		}
	}

	if l.cfg.HasHeader {
		if err := w.Write(cols); err != nil {
			return err
		}
	}

	for _, rec := range records {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		row := make([]string, len(cols))
		for i, col := range cols {
			row[i] = fmt.Sprintf("%v", rec[col])
		}
		if err := w.Write(row); err != nil {
			return err
		}
	}

	w.Flush()
	return w.Error()
}
