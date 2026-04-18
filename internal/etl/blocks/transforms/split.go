package transforms

import (
	"fmt"
	"strings"

	"github.com/rinjold/go-etl-studio/internal/etl/blocks"
	"github.com/rinjold/go-etl-studio/internal/etl/contracts"
)

func init() {
	blocks.Register("transform.split", func() contracts.Block { return &Split{} })
}

// Split divise le flux en N sorties selon des conditions.
// Paramètres :
//   condition.0 : condition pour le port de sortie 0 (ex: "country == FR")
//   condition.1 : condition pour le port de sortie 1 (ex: "country == DE")
//   condition.else : toutes les lignes qui ne matchent aucune condition (optionnel)
// Les ports de sortie sont ordonnés : out0, out1, ..., outElse
type Split struct{}

func (b *Split) Type() string { return "transform.split" }

func (b *Split) Run(bctx *contracts.BlockContext) error {
	if len(bctx.Inputs) == 0 {
		return fmt.Errorf("transform.split: aucun port d'entrée")
	}

	// Construire la liste des conditions dans l'ordre.
	type condPort struct {
		condition string
		portIdx   int
	}
	var conditions []condPort
	var elsePortIdx = -1

	for i := 0; i < len(bctx.Outputs); i++ {
		key := fmt.Sprintf("condition.%d", i)
		if cond, ok := bctx.Params[key]; ok {
			conditions = append(conditions, condPort{cond, i})
		}
	}
	if _, ok := bctx.Params["condition.else"]; ok {
		elsePortIdx = len(bctx.Outputs) - 1
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
			matched := false
			for _, cp := range conditions {
				parts := strings.Fields(cp.condition)
				if len(parts) == 3 && matchCondition(row, parts[0], parts[1], parts[2]) {
					if cp.portIdx < len(bctx.Outputs) {
						bctx.Outputs[cp.portIdx].Ch <- row
					}
					matched = true
					break
				}
			}
			if !matched && elsePortIdx >= 0 && elsePortIdx < len(bctx.Outputs) {
				bctx.Outputs[elsePortIdx].Ch <- row
			}
		}
	}
}
