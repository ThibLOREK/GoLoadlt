package transforms

import (
	"fmt"
	"strconv"
	"time"

	"github.com/rinjold/go-etl-studio/internal/etl/blocks"
	"github.com/rinjold/go-etl-studio/internal/etl/contracts"
)

func init() {
	blocks.Register("transform.cast", func() contracts.Block { return &Cast{} })
}

// Cast convertit le type d'une colonne.
// Paramètres :
//   column     : nom de la colonne à convertir
//   targetType : type cible ("int", "float", "bool", "string", "date")
//   format     : format de date si targetType=="date" (ex: "2006-01-02")
type Cast struct{}

func (b *Cast) Type() string { return "transform.cast" }

func (b *Cast) Run(bctx *contracts.BlockContext) error {
	col := bctx.Params["column"]
	targetType := bctx.Params["targetType"]
	if col == "" || targetType == "" {
		return fmt.Errorf("transform.cast: paramètres 'column' et 'targetType' obligatoires")
	}
	dateFormat := bctx.Params["format"]
	if dateFormat == "" {
		dateFormat = "2006-01-02"
	}

	if len(bctx.Inputs) == 0 {
		return fmt.Errorf("transform.cast: aucun port d'entrée")
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

			if raw, exists := newRow[col]; exists {
				str := fmt.Sprintf("%v", raw)
				switch targetType {
				case "int":
					if v, err := strconv.ParseInt(str, 10, 64); err == nil { newRow[col] = v }
				case "float":
					if v, err := strconv.ParseFloat(str, 64); err == nil { newRow[col] = v }
				case "bool":
					if v, err := strconv.ParseBool(str); err == nil { newRow[col] = v }
				case "string":
					newRow[col] = str
				case "date":
					if t, err := time.Parse(dateFormat, str); err == nil { newRow[col] = t }
				}
			}
			for _, out := range bctx.Outputs { out.Ch <- newRow }
		}
	}
}
