package transforms

import (
	"fmt"

	"github.com/rinjold/go-etl-studio/internal/etl/blocks"
	"github.com/rinjold/go-etl-studio/internal/etl/contracts"
	"github.com/rinjold/go-etl-studio/internal/etl/expression"
)

func init() {
	blocks.Register("transform.filter", func() contracts.Block { return &Filter{} })
}

// Filter conserve uniquement les lignes qui matchent une condition.
// Paramètres:
//   - condition : ex "amount > 100" ou "country == 'FR'"
type Filter struct{}

func (b *Filter) Type() string { return "transform.filter" }

func (b *Filter) Run(bctx *contracts.BlockContext) error {
	if len(bctx.Inputs) == 0 {
		return fmt.Errorf("transform.filter: aucun port d'entrée")
	}
	cond := bctx.Params["condition"]
	if cond == "" {
		return fmt.Errorf("transform.filter: paramètre 'condition' manquant")
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
			keep, err := expression.EvalBool(cond, row)
			if err != nil {
				return err
			}
			if !keep {
				continue
			}
			for _, out := range bctx.Outputs {
				select {
				case out.Ch <- row:
				case <-bctx.Ctx.Done():
					return bctx.Ctx.Err()
				}
			}
		}
	}
}
