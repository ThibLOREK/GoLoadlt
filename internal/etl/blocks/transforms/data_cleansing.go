package transforms

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"

	"github.com/rinjold/go-etl-studio/internal/etl/blocks"
	"github.com/rinjold/go-etl-studio/internal/etl/contracts"
)

func init() {
	blocks.Register("transform.data_cleansing", func() contracts.Block { return &DataCleansing{} })
}

// DataCleansing nettoie les données : trim, casse, suppression caractères spéciaux, nulls.
// Paramètres (tous optionnels, activés avec "true") :
//   trim            : supprime espaces début/fin
//   toLower / toUpper : normalise la casse
//   removeSpecial   : supprime les caractères non alphanumériques (hors espaces)
//   nullifyEmpty    : remplace les chaînes vides par nil
//   columns         : liste de colonnes séparées par virgule (défaut: toutes)
type DataCleansing struct{}

func (b *DataCleansing) Type() string { return "transform.data_cleansing" }

var reSpecial = regexp.MustCompile(`[^\p{L}\p{N}\s]`)

func (b *DataCleansing) Run(bctx *contracts.BlockContext) error {
	trim := bctx.Params["trim"] == "true"
	toLower := bctx.Params["toLower"] == "true"
	toUpper := bctx.Params["toUpper"] == "true"
	removeSpecial := bctx.Params["removeSpecial"] == "true"
	nullifyEmpty := bctx.Params["nullifyEmpty"] == "true"

	var targetCols map[string]bool
	if cols := bctx.Params["columns"]; cols != "" {
		targetCols = make(map[string]bool)
		for _, c := range strings.Split(cols, ",") {
			targetCols[strings.TrimSpace(c)] = true
		}
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
				if targetCols != nil && !targetCols[k] {
					newRow[k] = v
					continue
				}
				str := fmt.Sprintf("%v", v)
				if trim { str = strings.TrimFunc(str, unicode.IsSpace) }
				if toLower { str = strings.ToLower(str) }
				if toUpper { str = strings.ToUpper(str) }
				if removeSpecial { str = reSpecial.ReplaceAllString(str, "") }
				if nullifyEmpty && str == "" { newRow[k] = nil } else { newRow[k] = str }
			}
			for _, out := range bctx.Outputs { out.Ch <- newRow }
		}
	}
}