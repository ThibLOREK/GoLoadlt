package transforms

import (
	"fmt"

	"github.com/rinjold/go-etl-studio/internal/etl/blocks"
	"github.com/rinjold/go-etl-studio/internal/etl/contracts"
)

func init() {
	blocks.Register("transform.unpivot", func() contracts.Block { return &Unpivot{} })
}

// Unpivot transforme des colonnes en lignes.
// Params:
//   - columns  : colonnes à dépivoter, séparées par virgule (ex: "jan,fev,mar")
//   - keyName  : nom de la colonne clé produite (ex: "mois")
//   - valueName: nom de la colonne valeur produite (ex: "montant")
type Unpivot struct{}

func (b *Unpivot) Type() string { return "transform.unpivot" }

func (b *Unpivot) Run(bctx *contracts.BlockContext) error {
	colsCsv := bctx.Params["columns"]
	keyName := bctx.Params["keyName"]
	valueName := bctx.Params["valueName"]
	if colsCsv == "" || keyName == "" || valueName == "" {
		return fmt.Errorf("transform.unpivot: paramètres 'columns', 'keyName', 'valueName' requis")
	}
	if len(bctx.Inputs) == 0 {
		return fmt.Errorf("transform.unpivot: aucun port d'entrée")
	}

	cols := splitComma(colsCsv)
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
			// Pour chaque colonne à dépivoter, émettre une ligne.
			for _, col := range cols {
				newRow := make(contracts.DataRow, len(row)-len(cols)+2)
				for k, v := range row {
					skip := false
					for _, c := range cols {
						if k == c {
							skip = true
							break
						}
					}
					if !skip {
						newRow[k] = v
					}
				}
				newRow[keyName] = col
				newRow[valueName] = row[col]
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
}
