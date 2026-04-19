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
// Utile pour observer et vérifier les données en transit dans un pipeline
// (l'équivalent du bloc "Dummy" de Pentaho PDI).
type Dummy struct{}

func (b *Dummy) Type() string { return "transform.dummy" }

func (b *Dummy) Run(bctx *contracts.BlockContext) error {
	if len(bctx.Inputs) == 0 {
		return fmt.Errorf("transform.dummy: aucun port d'entrée")
	}
	in := bctx.Inputs[0]

	closeOutputs := func() {
		for _, out := range bctx.Outputs {
			close(out.Ch)
		}
	}

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
