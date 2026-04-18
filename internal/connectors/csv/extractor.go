package csv

import (
	"context"
	"encoding/csv"
	"io"
	"os"

	"github.com/rinjold/go-etl-studio/internal/etl/contracts"
)

type ExtractorConfig struct {
	FilePath  string
	Delimiter rune
	HasHeader bool
}

type Extractor struct {
	cfg ExtractorConfig
}

func NewExtractor(cfg ExtractorConfig) *Extractor {
	if cfg.Delimiter == 0 {
		cfg.Delimiter = ','
	}
	return &Extractor{cfg: cfg}
}

func (e *Extractor) Extract(ctx context.Context) ([]contracts.Record, error) {
	f, err := os.Open(e.cfg.FilePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	r := csv.NewReader(f)
	r.Comma = e.cfg.Delimiter

	var headers []string
	if e.cfg.HasHeader {
		headers, err = r.Read()
		if err != nil {
			return nil, err
		}
	}

	var records []contracts.Record
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		row, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		rec := make(contracts.Record, len(row))
		for i, val := range row {
			key := columnKey(headers, i)
			rec[key] = val
		}
		records = append(records, rec)
	}

	return records, nil
}

func columnKey(headers []string, i int) string {
	if i < len(headers) {
		return headers[i]
	}
	return string(rune('A' + i))
}
