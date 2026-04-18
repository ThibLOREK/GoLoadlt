package postgres

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rinjold/go-etl-studio/internal/etl/contracts"
)

type LoaderConfig struct {
	Pool      *pgxpool.Pool
	Schema    string
	TableName string
	Columns   []string
	BatchSize int
}

type Loader struct {
	cfg LoaderConfig
}

func NewLoader(cfg LoaderConfig) *Loader {
	if cfg.BatchSize == 0 {
		cfg.BatchSize = 500
	}
	if cfg.Schema == "" {
		cfg.Schema = "public"
	}
	return &Loader{cfg: cfg}
}

func (l *Loader) Load(ctx context.Context, records []contracts.Record) error {
	if len(records) == 0 {
		return nil
	}

	cols := l.cfg.Columns
	if len(cols) == 0 {
		for k := range records[0] {
			cols = append(cols, k)
		}
	}

	table := fmt.Sprintf("%s.%s", l.cfg.Schema, l.cfg.TableName)
	placeholders := make([]string, len(cols))
	for i := range cols {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
	}

	q := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s)",
		table,
		strings.Join(cols, ", "),
		strings.Join(placeholders, ", "),
	)

	for i := 0; i < len(records); i += l.cfg.BatchSize {
		end := i + l.cfg.BatchSize
		if end > len(records) {
			end = len(records)
		}
		batch := records[i:end]

		tx, err := l.cfg.Pool.Begin(ctx)
		if err != nil {
			return err
		}

		for _, rec := range batch {
			args := make([]any, len(cols))
			for j, col := range cols {
				args[j] = rec[col]
			}
			if _, err := tx.Exec(ctx, q, args...); err != nil {
				_ = tx.Rollback(ctx)
				return err
			}
		}

		if err := tx.Commit(ctx); err != nil {
			return err
		}
	}

	return nil
}
