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

// Join joint deux flux (left = Inputs[0], right = Inputs[1]) sur une clé.
// Paramètres:
//   - leftKey  : colonne clé du flux gauche  (ex: "user_id")
//   - rightKey : colonne clé du flux droit   (ex: "id")
//   - type     : inner | left | right | full  (défaut: inner)
//
// Stratégie : hash-join — le flux droit est chargé en mémoire (build phase),
// puis le flux gauche est streamé ligne par ligne (probe phase).
type Join struct{}

func (b *Join) Type() string { return "transform.join" }

func (b *Join) Run(bctx *contracts.BlockContext) error {
	if len(bctx.Inputs) < 2 {
		return fmt.Errorf("transform.join: 2 ports d'entrée requis (left, right)")
	}
	leftKey := bctx.Params["leftKey"]
	rightKey := bctx.Params["rightKey"]
	joinType := strings.ToLower(bctx.Params["type"])
	if joinType == "" {
		joinType = "inner"
	}
	if leftKey == "" || rightKey == "" {
		return fmt.Errorf("transform.join: params 'leftKey' et 'rightKey' requis")
	}
	switch joinType {
	case "inner", "left", "right", "full":
	default:
		return fmt.Errorf("transform.join: type '%s' non supporté (inner|left|right|full)", joinType)
	}

	// --- Phase build : charger le flux droit en mémoire ---
	rightMap := make(map[string][]contracts.DataRow)
	for row := range bctx.Inputs[1].Ch {
		k := fmt.Sprintf("%v", row[rightKey])
		rightMap[k] = append(rightMap[k], row)
	}

	// Pour full join : on track les clés droites émises.
	emittedRight := make(map[string]bool)

	close := func() {
		for _, out := range bctx.Outputs {
			close(out.Ch)
		}
	}

	emit := func(row contracts.DataRow) error {
		for _, out := range bctx.Outputs {
			select {
			case out.Ch <- row:
			case <-bctx.Ctx.Done():
				return bctx.Ctx.Err()
			}
		}
		return nil
	}

	// --- Phase probe : streamé sur le flux gauche ---
	for {
		select {
		case <-bctx.Ctx.Done():
			close()
			return bctx.Ctx.Err()
		case leftRow, ok := <-bctx.Inputs[0].Ch:
			if !ok {
				// Pour full join : émettre les lignes droites non matchées.
				if joinType == "full" || joinType == "right" {
					for k, rows := range rightMap {
						if emittedRight[k] {
							continue
						}
						for _, r := range rows {
							if err := emit(r); err != nil {
								return err
							}
						}
					}
				}
				close()
				return nil
			}

			k := fmt.Sprintf("%v", leftRow[leftKey])
			rightRows, matched := rightMap[k]

			if matched {
				emittedRight[k] = true
				for _, rightRow := range rightRows {
					merged := mergeRows(leftRow, rightRow, rightKey)
					if err := emit(merged); err != nil {
						return err
					}
				}
			} else if joinType == "left" || joinType == "full" {
				// Ligne gauche sans correspondance.
				if err := emit(leftRow); err != nil {
					return err
				}
			}
		}
	}
}

// mergeRows fusionne deux lignes en préfixant les clés droites dupliquées de "right_".
func mergeRows(left, right contracts.DataRow, rightKey string) contracts.DataRow {
	merged := make(contracts.DataRow, len(left)+len(right))
	for k, v := range left {
		merged[k] = v
	}
	for k, v := range right {
		if k == rightKey {
			continue // la clé droite est déjà dans left sous leftKey
		}
		if _, exists := merged[k]; exists {
			merged["right_"+k] = v
		} else {
			merged[k] = v
		}
	}
	return merged
}
