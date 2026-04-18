package transforms

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/rinjold/go-etl-studio/internal/etl/blocks"
	"github.com/rinjold/go-etl-studio/internal/etl/contracts"
)

func init() {
	blocks.Register("transform.filter", func() contracts.Block { return &Filter{} })
}

// Filter ne laisse passer que les lignes qui satisfont une condition simple.
// Syntaxe supportée : "colonne operateur valeur"
// Exemples : "amount > 100", "status == active", "region != FR"
type Filter struct{}

func (b *Filter) Type() string { return "transform.filter" }

func (b *Filter) Run(bctx *contracts.BlockContext) error {
	condition := bctx.Params["condition"]
	if condition == "" {
		return fmt.Errorf("transform.filter: paramètre 'condition' manquant")
	}

	parts := strings.Fields(condition)
	if len(parts) != 3 {
		return fmt.Errorf("transform.filter: condition invalide '%s' (attendu: 'col op val')", condition)
	}
	col, op, val := parts[0], parts[1], parts[2]

	if len(bctx.Inputs) == 0 {
		return fmt.Errorf("transform.filter: aucun port d'entrée")
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
			if matchCondition(row, col, op, val) {
				for _, out := range bctx.Outputs {
					out.Ch <- row
				}
			}
		}
	}
}

func matchCondition(row contracts.DataRow, col, op, val string) bool {
	rawVal, ok := row[col]
	if !ok {
		return false
	}
	cellStr := fmt.Sprintf("%v", rawVal)

	// Comparaison numérique si possible.
	cellNum, errCell := strconv.ParseFloat(cellStr, 64)
	valNum, errVal := strconv.ParseFloat(val, 64)

	if errCell == nil && errVal == nil {
		switch op {
		case ">": return cellNum > valNum
		case ">=": return cellNum >= valNum
		case "<": return cellNum < valNum
		case "<=": return cellNum <= valNum
		case "==", "=": return cellNum == valNum
		case "!=": return cellNum != valNum
		}
	}

	// Comparaison string.
	switch op {
	case "==", "=": return cellStr == val
	case "!=": return cellStr != val
	case "contains": return strings.Contains(cellStr, val)
	case "startsWith": return strings.HasPrefix(cellStr, val)
	case "endsWith": return strings.HasSuffix(cellStr, val)
	}
	return false
}
