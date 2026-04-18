package transforms

import (
	"fmt"
	"strings"

	"github.com/rinjold/go-etl-studio/internal/etl/blocks"
	"github.com/rinjold/go-etl-studio/internal/etl/contracts"
)

func init() {
	blocks.Register("transform.dedup", func() contracts.Block { return &Dedup{} })
}

// Dedup supprime les doublons selon les colonnes clés.
// Params:
//   - keys : colonnes clés séparées par virgule (ex: "id,email")
type Dedup struct{}

func (b *Dedup) Type() string { return "transform.dedup" }

func (b *Dedup) Run(bctx *contracts.BlockContext) error {
	keysCSV := bctx.Params["keys"]
	if keysCSV == "" {
		return fmt.Errorf("transform.dedup: paramètre 'keys' manquant")
	}
	if len(bctx.Inputs) == 0 {
		return fmt.Errorf("transform.dedup: aucun port d'entrée")
	}
	keys := splitComma(keysCSV)
	seen := make(map[string]bool)
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
			// Construire la clé composite.
			parts := make([]string, len(keys))
			for i, k := range keys {
				parts[i] = fmt.Sprintf("%v", row[k])
			}
			composite := strings.Join(parts, "||")
			if seen[composite] {
				continue
			}
			seen[composite] = true
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
