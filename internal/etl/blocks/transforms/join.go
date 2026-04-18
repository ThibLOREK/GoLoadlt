package transforms

import (
	"fmt"
	"strings"

	"github.com/rinjold/go-etl-studio/internal/etl/blocks"
	"github.com/rinjold/go-etl-studio/internal/etl/contracts"
)

func init() {
	blocks.Register("transform.join", func() contracts.Block { return &Join{} })
}

// Join effectue une jointure entre deux flux.
// Paramètres :
//   leftKey   : colonne clé du flux gauche
//   rightKey  : colonne clé du flux droit
//   type      : inner (défaut) | left | right
type Join struct{}

func (b *Join) Type() string { return "transform.join" }

func (b *Join) Run(bctx *contracts.BlockContext) error {
	if len(bctx.Inputs) < 2 {
		return fmt.Errorf("transform.join: 2 ports d'entrée requis (left, right)")
	}
	leftKey := bctx.Params["leftKey"]
	rightKey := bctx.Params["rightKey"]
	joinType := bctx.Params["type"]
	if joinType == "" { joinType = "inner" }
	if leftKey == "" || rightKey == "" {
		return fmt.Errorf("transform.join: paramètres 'leftKey' et 'rightKey' obligatoires")
	}

	// Charger le flux droit en mémoire (build side).
	rightIndex := make(map[string][]contracts.DataRow)
	for row := range bctx.Inputs[1].Ch {
		key := fmt.Sprintf("%v", row[rightKey])
		rightIndex[key] = append(rightIndex[key], row)
	}

	// Streamer le flux gauche et joindre.
	for leftRow := range bctx.Inputs[0].Ch {
		key := fmt.Sprintf("%v", leftRow[leftKey])
		rightRows, found := rightIndex[key]

		if found {
			for _, rightRow := range rightRows {
				merged := make(contracts.DataRow, len(leftRow)+len(rightRow))
				for k, v := range leftRow { merged[k] = v }
				for k, v := range rightRow {
					colName := k
					if _, exists := leftRow[k]; exists && k != rightKey {
						colName = "right_" + k
					}
					merged[colName] = v
				}
				for _, out := range bctx.Outputs { out.Ch <- merged }
			}
		} else if strings.ToLower(joinType) == "left" {
			for _, out := range bctx.Outputs { out.Ch <- leftRow }
		}
	}
	for _, out := range bctx.Outputs { close(out.Ch) }
	return nil
}