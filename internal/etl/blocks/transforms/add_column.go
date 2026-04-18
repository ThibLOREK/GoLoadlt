package transforms

import (
	"fmt"

	"github.com/rinjold/go-etl-studio/internal/etl/blocks"
	"github.com/rinjold/go-etl-studio/internal/etl/contracts"
	"github.com/rinjold/go-etl-studio/internal/etl/expression"
)

func init() {
	blocks.Register("transform.add_column", func() contracts.Block { return &AddColumn{} })
}

// AddColumn ajoute une colonne calculée à partir d'une expression simple.
// Paramètres:
//   - name       : nom de la nouvelle colonne
//   - expression : ex "amount * 1.2" ou "country"
type AddColumn struct{}

func (b *AddColumn) Type() string { return "transform.add_column" }

func (b *AddColumn) Run(bctx *contracts.BlockContext) error {
	if len(bctx.Inputs) == 0 {
		return fmt.Errorf("transform.add_column: aucun port d'entrée")
	}
	name := bctx.Params["name"]
	expr := bctx.Params["expression"]
	if name == "" || expr == "" {
		return fmt.Errorf("transform.add_column: paramètres 'name' et 'expression' requis")
	}

	in := bctx.Inputs[0]
	for {
		select {
		case <-bctx.Ctx.Done():
			for _, out := range bctx.Outputs {
				close(out.Ch)
			}
			return bctx.Ctx.Err()
		case row, ok := <-in.Ch:
			if !ok {
				for _, out := range bctx.Outputs {
					close(out.Ch)
				}
				return nil
			}
			val, err := expression.EvalValue(expr, row)
			if err != nil {
				return err
			}
			newRow := make(contracts.DataRow, len(row)+1)
			for k, v := range row {
				newRow[k] = v
			}
			newRow[name] = val
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
