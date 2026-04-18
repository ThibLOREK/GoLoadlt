package targets

import (
	"encoding/json"
	"fmt"

	"github.com/rinjold/go-etl-studio/internal/etl/blocks"
	"github.com/rinjold/go-etl-studio/internal/etl/contracts"
)

func init() {
	blocks.Register("target.browse", func() contracts.Block { return &BrowseTarget{} })
}

// BrowseTarget collecte les lignes en mémoire pour les afficher dans l'UI (aperçu).
// Les lignes sont stockées dans Rows et accessibles après l'exécution.
type BrowseTarget struct {
	Rows []contracts.DataRow
	Limit int
}

func (b *BrowseTarget) Type() string { return "target.browse" }

func (b *BrowseTarget) Run(bctx *contracts.BlockContext) error {
	if len(bctx.Inputs) == 0 {
		return fmt.Errorf("target.browse: aucun port d'entrée")
	}
	limit := 1000
	if b.Limit > 0 {
		limit = b.Limit
	}
	in := bctx.Inputs[0]
	for {
		select {
		case <-bctx.Ctx.Done():
			return bctx.Ctx.Err()
		case row, ok := <-in.Ch:
			if !ok {
				return nil
			}
			if len(b.Rows) < limit {
				b.Rows = append(b.Rows, row)
			}
		}
	}
}

// ToJSON retourne les lignes collectées en JSON (pour l'API).
func (b *BrowseTarget) ToJSON() ([]byte, error) {
	return json.Marshal(b.Rows)
}
