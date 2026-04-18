package transformers

import (
	"context"

	"github.com/rinjold/go-etl-studio/internal/etl/contracts"
)

// Mapper renomme les colonnes selon un mapping source → cible.
type Mapper struct {
	Mapping map[string]string // {"old_col": "new_col"}
}

func (m Mapper) Transform(ctx context.Context, in []contracts.Record) ([]contracts.Record, error) {
	out := make([]contracts.Record, 0, len(in))
	for _, rec := range in {
		newRec := make(contracts.Record, len(rec))
		for k, v := range rec {
			if newKey, ok := m.Mapping[k]; ok {
				newRec[newKey] = v
			} else {
				newRec[k] = v
			}
		}
		out = append(out, newRec)
	}
	return out, nil
}
