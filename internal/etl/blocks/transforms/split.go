package transforms

import (
	"fmt"

	"github.com/rinjold/go-etl-studio/internal/etl/blocks"
	"github.com/rinjold/go-etl-studio/internal/etl/contracts"
	"github.com/rinjold/go-etl-studio/internal/etl/expression"
)

func init() {
	blocks.Register("transform.split", func() contracts.Block { return &Split{} })
}

// Split route chaque ligne vers la première sortie dont la condition est vraie.
// La dernière sortie est le "else" (catch-all).
// Paramètres:
//   - conditions : liste CSV de conditions correspondant aux sorties 0..N-2
//                  ex : "amount > 1000, amount > 500"
type Split struct{}

func (b *Split) Type() string { return "transform.split" }

func (b *Split) Run(bctx *contracts.BlockContext) error {
	if len(bctx.Inputs) == 0 {
		return fmt.Errorf("transform.split: aucun port d'entrée")
	}
	if len(bctx.Outputs) < 2 {
		return fmt.Errorf("transform.split: au moins 2 ports de sortie requis")
	}
	condsRaw := bctx.Params["conditions"]
	if condsRaw == "" {
		return fmt.Errorf("transform.split: paramètre 'conditions' manquant")
	}
	conds := splitComma(condsRaw)
	if len(conds) != len(bctx.Outputs)-1 {
		return fmt.Errorf(
			"transform.split: %d conditions mais %d sorties (attendu %d conditions)",
			len(conds), len(bctx.Outputs), len(bctx.Outputs)-1,
		)
	}

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
			routed := false
			for i, cond := range conds {
				match, err := expression.EvalBool(cond, row)
				if err != nil {
					return err
				}
				if match {
					select {
					case bctx.Outputs[i].Ch <- row:
					case <-bctx.Ctx.Done():
						return bctx.Ctx.Err()
					}
					routed = true
					break
				}
			}
			if !routed {
				// Catch-all : dernière sortie.
				select {
				case bctx.Outputs[len(bctx.Outputs)-1].Ch <- row:
				case <-bctx.Ctx.Done():
					return bctx.Ctx.Err()
				}
			}
		}
	}
}
