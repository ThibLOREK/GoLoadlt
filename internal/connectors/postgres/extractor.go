package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rinjold/go-etl-studio/internal/etl/contracts"
)

type ExtractorConfig struct {
	Pool      *pgxpool.Pool
	Schema    string
	TableName string
	Columns   []string
	Where     string
	BatchSize int
}

type Extractor struct {
	cfg ExtractorConfig
}

func NewExtractor(cfg ExtractorConfig) *Extractor {
	if cfg.BatchSize == 0 {
		cfg.BatchSize = 1000
	}
	if cfg.Schema == "" {
		cfg.Schema = "public"
	}
	return &Extractor{cfg: cfg}
}

func (e *Extractor) Extract(ctx context.Context) ([]contracts.Record, error) {
	cols := "*"
	if len(e.cfg.Columns) > 0 {
		cols = ""
		for i, c := range e.cfg.Columns {
			if i > 0 {
				cols += ", "
			}
			cols += c
		}
	}

	table := fmt.Sprintf("%s.%s", e.cfg.Schema, e.cfg.TableName)
	q := fmt.Sprintf("SELECT %s FROM %s", cols, table)
	if e.cfg.Where != "" {
		q += " WHERE " + e.cfg.Where
	}

	rows, err := e.cfg.Pool.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	fieldDescriptions := rows.FieldDescriptions()
	headers := make([]string, len(fieldDescriptions))
	for i, fd := range fieldDescriptions {
		headers[i] = string(fd.Name)
	}

	records := make([]contracts.Record, 0, e.cfg.BatchSize)
	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			return nil, err
		}
		rec := make(contracts.Record, len(headers))
		for i, h := range headers {
			rec[h] = values[i]
		}
		records = append(records, rec)
	}

	return records, rows.Err()
}
