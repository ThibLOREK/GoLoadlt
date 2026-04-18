package transforms

import (
	"fmt"

	"github.com/rinjold/go-etl-studio/internal/etl/blocks"
	"github.com/rinjold/go-etl-studio/internal/etl/contracts"
)

func init() {
	blocks.Register("transform.select", func() contracts.Block { return &Select{} })
}

// Select ne conserve qu'un sous-ensemble de colonnes.
// Paramètres:
//   - columns : liste de colonnes séparées par des virgules

type Select struct{}

func (b *Select) Type() string { return "transform.select" }

func (b *Select) Run(bctx *contracts.BlockContext) error {
	if len(bctx.Inputs) == 0 {
		return fmt.Errorf("transform.select: aucun port d'entrée")
	}
	colsRaw := bctx.Params["columns"]
	if colsRaw == "" {
		return fmt.Errorf("transform.select: paramètre 'columns' manquant")
	}
	cols := splitComma(colsRaw)
	in := bctx.Inputs[0]

	for {
		select {
		case <-bctx.Ctx.Done():
			for _, out := range bctx.Outputs { close(out.Ch) }
			return bctx.Ctx.Err()
		case row, ok := <-in.Ch:
			if !ok {
				for _, out := range bctx.Outputs { close(out.Ch) }
				return nil
			}
			newRow := make(contracts.DataRow, len(cols))
			for _, col := range cols {
				newRow[col] = row[col]
			}
			for _, out := range bctx.Outputs {
				select {
				case out.Ch <- newRow:
				case <-bctx.Ctx.Done():
					return bctx.Ctx.Err()
				}
			}
		}
	}
}
