package transforms

import (
	"fmt"
	"strconv"

	"github.com/rinjold/go-etl-studio/internal/etl/blocks"
	"github.com/rinjold/go-etl-studio/internal/etl/contracts"
)

func init() {
	blocks.Register("transform.cast", func() contracts.Block { return &Cast{} })
}

// Cast convertit le type d'une colonne.
// Paramètres:
//   - column     : colonne à caster
//   - targetType : string | int | float | bool

type Cast struct{}

func (b *Cast) Type() string { return "transform.cast" }

func (b *Cast) Run(bctx *contracts.BlockContext) error {
	if len(bctx.Inputs) == 0 {
		return fmt.Errorf("transform.cast: aucun port d'entrée")
	}
	col := bctx.Params["column"]
	targetType := bctx.Params["targetType"]
	if col == "" || targetType == "" {
		return fmt.Errorf("transform.cast: paramètres 'column' et 'targetType' requis")
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
			for k, v := range row {
				newRow[k] = v
			}
			casted, err := castValue(row[col], targetType)
			if err != nil {
				return err
			}
			newRow[col] = casted
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

func castValue(v any, targetType string) (any, error) {
	s := fmt.Sprintf("%v", v)
	switch targetType {
	case "string":
		return s, nil
	case "int":
		return strconv.Atoi(s)
	case "float":
		return strconv.ParseFloat(s, 64)
	case "bool":
		return strconv.ParseBool(s)
	default:
		return nil, fmt.Errorf("transform.cast: type cible non supporté: %s", targetType)
	}
}
