package transformers

import (
	"context"

	"github.com/rinjold/go-etl-studio/internal/etl/contracts"
)

type Noop struct{}

func (n Noop) Transform(ctx context.Context, in []contracts.Record) ([]contracts.Record, error) {
	return in, nil
}
