package transforms

import (
	"fmt"
	"strings"

	"github.com/rinjold/go-etl-studio/internal/etl/blocks"
	"github.com/rinjold/go-etl-studio/internal/etl/contracts"
)

func init() {
	blocks.Register("transform.find_replace", func() contracts.Block { return &FindReplace{} })
}

// FindReplace cherche et remplace des valeurs dans une colonne.
// Paramètres :
//   column  : colonne cible
//   find    : valeur à chercher
//   replace : valeur de remplacement
//   mode    : exact (défaut) | contains | regex
type FindReplace struct{}

func (b *FindReplace) Type() string { return "transform.find_replace" }

func (b *FindReplace) Run(bctx *contracts.BlockContext) error {
	col := bctx.Params["column"]
	find := bctx.Params["find"]
	replace := bctx.Params["replace"]
	if col == "" || find == "" {
		return fmt.Errorf("transform.find_replace: 'column' et 'find' obligatoires")
	}
	if len(bctx.Inputs) == 0 {
		return fmt.Errorf("transform.find_replace: aucun port d'entrée")
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
			newRow := make(contracts.DataRow, len(row))
			for k, v := range row { newRow[k] = v }
			if val, exists := newRow[col]; exists {
				str := fmt.Sprintf("%v", val)
				newRow[col] = strings.ReplaceAll(str, find, replace)
			}
			for _, out := range bctx.Outputs { out.Ch <- newRow }
		}
	}
}