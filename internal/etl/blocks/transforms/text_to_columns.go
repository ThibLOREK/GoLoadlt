package transforms

import (
	"fmt"
	"strings"

	"github.com/rinjold/go-etl-studio/internal/etl/blocks"
	"github.com/rinjold/go-etl-studio/internal/etl/contracts"
)

func init() {
	blocks.Register("transform.text_to_columns", func() contracts.Block { return &TextToColumns{} })
}

// TextToColumns découpe une colonne texte en plusieurs colonnes.
// Paramètres :
//   column    : colonne à découper
//   delimiter : délimiteur (défaut: ",")
//   prefix    : préfixe des colonnes générées (défaut: column+"_")
//   maxSplit  : nombre max de colonnes (défaut: 0 = illimité)
type TextToColumns struct{}

func (b *TextToColumns) Type() string { return "transform.text_to_columns" }

func (b *TextToColumns) Run(bctx *contracts.BlockContext) error {
	col := bctx.Params["column"]
	delim := bctx.Params["delimiter"]
	prefix := bctx.Params["prefix"]
	if col == "" { return fmt.Errorf("transform.text_to_columns: 'column' obligatoire") }
	if delim == "" { delim = "," }
	if prefix == "" { prefix = col + "_" }

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
			str := fmt.Sprintf("%v", row[col])
			parts := strings.Split(str, delim)
			for i, p := range parts {
				newRow[fmt.Sprintf("%s%d", prefix, i+1)] = strings.TrimSpace(p)
			}
			for _, out := range bctx.Outputs { out.Ch <- newRow }
		}
	}
}