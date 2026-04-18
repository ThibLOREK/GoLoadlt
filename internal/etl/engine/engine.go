package engine

import (
	"context"

	"github.com/rinjold/go-etl-studio/internal/etl/contracts"
)

type Engine struct {
	Extractor    contracts.Extractor
	Transformer  contracts.Transformer
	Loader       contracts.Loader
}

func (e Engine) Run(ctx context.Context) error {
	records, err := e.Extractor.Extract(ctx)
	if err != nil {
		return err
	}

	if e.Transformer != nil {
		records, err = e.Transformer.Transform(ctx, records)
		if err != nil {
			return err
		}
	}

	return e.Loader.Load(ctx, records)
}
