package transforms

import (
	"fmt"
	"sort"
	"strconv"

	"github.com/rinjold/go-etl-studio/internal/etl/blocks"
	"github.com/rinjold/go-etl-studio/internal/etl/contracts"
)

func init() {
	blocks.Register("transform.sort", func() contracts.Block { return &Sort{} })
}

// Sort trie les lignes selon une ou plusieurs colonnes.
// Params:
//   - columns : colonnes de tri séparées par virgule (ex: "date,amount")
//   - order   : "asc" (défaut) ou "desc"
type Sort struct{}

func (b *Sort) Type() string { return "transform.sort" }

func (b *Sort) Run(bctx *contracts.BlockContext) error {
	colsCSV := bctx.Params["columns"]
	if colsCSV == "" {
		return fmt.Errorf("transform.sort: paramètre 'columns' manquant")
	}
	if len(bctx.Inputs) == 0 {
		return fmt.Errorf("transform.sort: aucun port d'entrée")
	}
	cols := splitComma(colsCSV)
	desc := bctx.Params["order"] == "desc"

	// Collecter toutes les lignes (tri nécessite la mémoire complète).
	var rows []contracts.DataRow
	for row := range bctx.Inputs[0].Ch {
		rows = append(rows, row)
	}

	sort.SliceStable(rows, func(i, j int) bool {
		for _, col := range cols {
			a := fmt.Sprintf("%v", rows[i][col])
			b2 := fmt.Sprintf("%v", rows[j][col])
			// Comparaison numérique si possible.
			an, ea := strconv.ParseFloat(a, 64)
			bn, eb := strconv.ParseFloat(b2, 64)
			var less bool
			if ea == nil && eb == nil {
				less = an < bn
			} else {
				less = a < b2
			}
			if a != b2 {
				if desc {
					return !less
				}
				return less
			}
		}
		return false
	})

	for _, row := range rows {
		for _, out := range bctx.Outputs {
			select {
			case out.Ch <- row:
			case <-bctx.Ctx.Done():
				return bctx.Ctx.Err()
			}
		}
	}
	for _, out := range bctx.Outputs {
		close(out.Ch)
	}
	return nil
}
