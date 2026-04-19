package transforms

import (
	"fmt"

	"github.com/rinjold/go-etl-studio/internal/etl/blocks"
	"github.com/rinjold/go-etl-studio/internal/etl/contracts"
)

func init() {
	blocks.Register("transform.dummy", func() contracts.Block { return &Dummy{} })
}

// Dummy laisse passer toutes les lignes sans aucune modification.
// Utile pour observer les données en transit (équivalent Pentaho PDI Dummy).
// Tolère 0 sortie (bloc terminal d'observation sans suite dans le pipeline).
type Dummy struct{}

func (b *Dummy) Type() string { return "transform.dummy" }

func (b *Dummy) Run(bctx *contracts.BlockContext) error {
	if len(bctx.Inputs) == 0 {
		return fmt.Errorf("transform.dummy: aucun port d'entrée")
	}

	closeOutputs := func() {
		for _, out := range bctx.Outputs {
			close(out.Ch)
		}
	}

	in := bctx.Inputs[0]
	for {
		select {
		case <-bctx.Ctx.Done():
			closeOutputs()
			return bctx.Ctx.Err()
		case row, ok := <-in.Ch:
			if !ok {
				closeOutputs()
				return nil
			}
			// Propager vers toutes les sorties (0 à N sorties supportées).
			for _, out := range bctx.Outputs {
				select {
				case out.Ch <- row:
				case <-bctx.Ctx.Done():
					closeOutputs()
					return bctx.Ctx.Err()
				}
			}
		}
	}
}
