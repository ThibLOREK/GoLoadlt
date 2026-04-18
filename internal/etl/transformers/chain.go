package transformers

import (
	"context"

	"github.com/rinjold/go-etl-studio/internal/etl/contracts"
)

// Chain enchaîne plusieurs transformers séquentiellement.
type Chain struct {
	Steps []contracts.Transformer
}

func (c Chain) Transform(ctx context.Context, in []contracts.Record) ([]contracts.Record, error) {
	var err error
	records := in
	for _, step := range c.Steps {
		records, err = step.Transform(ctx, records)
		if err != nil {
			return nil, err
		}
	}
	return records, nil
}
