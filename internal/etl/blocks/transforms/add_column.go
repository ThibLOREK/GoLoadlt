package transforms

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/rinjold/go-etl-studio/internal/etl/blocks"
	"github.com/rinjold/go-etl-studio/internal/etl/contracts"
)

func init() {
	blocks.Register("transform.add_column", func() contracts.Block { return &AddColumn{} })
}

// AddColumn ajoute une colonne calculée à chaque ligne.
// Paramètres :
//   name       : nom de la nouvelle colonne
//   expression : expression simple, ex: "price * quantity", "'FR'", "col1"
// Expressions supportées :
//   - référence à une colonne existante : "colName"
//   - valeur littérale string entre guillemets simples : "'valeur'"
//   - opération arithmétique : "col1 * col2", "col1 + 10"
type AddColumn struct{}

func (b *AddColumn) Type() string { return "transform.add_column" }

func (b *AddColumn) Run(bctx *contracts.BlockContext) error {
	name := bctx.Params["name"]
	expr := bctx.Params["expression"]
	if name == "" || expr == "" {
		return fmt.Errorf("transform.add_column: paramètres 'name' et 'expression' obligatoires")
	}

	if len(bctx.Inputs) == 0 {
		return fmt.Errorf("transform.add_column: aucun port d'entrée")
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
			newRow := make(contracts.DataRow, len(row)+1)
			for k, v := range row { newRow[k] = v }
			newRow[name] = evalExpression(expr, row)
			for _, out := range bctx.Outputs { out.Ch <- newRow }
		}
	}
}

// evalExpression évalue une expression simple.
func evalExpression(expr string, row contracts.DataRow) any {
	expr = strings.TrimSpace(expr)

	// Littéral string : 'valeur'
	if strings.HasPrefix(expr, "'") && strings.HasSuffix(expr, "'") {
		return expr[1 : len(expr)-1]
	}

	// Opération arithmétique : col1 op col2 ou col1 op num
	for _, op := range []string{" * ", " / ", " + ", " - "} {
		if idx := strings.Index(expr, op); idx >= 0 {
			left := strings.TrimSpace(expr[:idx])
			right := strings.TrimSpace(expr[idx+len(op):])
			leftVal := resolveNumeric(left, row)
			rightVal := resolveNumeric(right, row)
			switch strings.TrimSpace(op) {
			case "*": return leftVal * rightVal
			case "/":
				if rightVal == 0 { return nil }
				return leftVal / rightVal
			case "+": return leftVal + rightVal
			case "-": return leftVal - rightVal
			}
		}
	}

	// Référence à une colonne.
	if val, ok := row[expr]; ok {
		return val
	}

	// Littéral numérique.
	if f, err := strconv.ParseFloat(expr, 64); err == nil {
		return f
	}

	return expr
}

func resolveNumeric(token string, row contracts.DataRow) float64 {
	if val, ok := row[token]; ok {
		if f, err := strconv.ParseFloat(fmt.Sprintf("%v", val), 64); err == nil {
			return f
		}
	}
	if f, err := strconv.ParseFloat(token, 64); err == nil {
		return f
	}
	return 0
}
