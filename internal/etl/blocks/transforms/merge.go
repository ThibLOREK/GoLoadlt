package transforms

import (
	"fmt"
	"strings"

	"github.com/rinjold/go-etl-studio/internal/etl/blocks"
	"github.com/rinjold/go-etl-studio/internal/etl/contracts"
)

func init() {
	blocks.Register("transform.merge", func() contracts.Block { return &Merge{} })
}

// Merge joint deux flux avec une parité complète pd.merge / df.merge.
//
// Paramètres (bctx.Params) :
//   - how        : "inner" | "left" | "right" | "outer"  (défaut: "inner")
//   - on         : colonne commune (si même nom des deux côtés), ex: "id"
//   - left_on    : colonne clé du flux gauche  (prioritaire sur "on")
//   - right_on   : colonne clé du flux droit   (prioritaire sur "on")
//   - left_suffix  : suffixe colonnes gauche en cas de collision (défaut: "_x")
//   - right_suffix : suffixe colonnes droite en cas de collision (défaut: "_y")
//   - validate   : "" | "one_to_one" | "one_to_many" | "many_to_one" | "many_to_many"
//
// Stratégie : hash-join — le flux droit est chargé en mémoire (build phase),
// puis le flux gauche est streamé ligne par ligne (probe phase).
type Merge struct{}

func (b *Merge) Type() string { return "transform.merge" }

func (b *Merge) Run(bctx *contracts.BlockContext) error {
	if len(bctx.Inputs) < 2 {
		return fmt.Errorf("transform.merge: 2 ports d'entrée requis (left, right)")
	}

	// --- Résolution des clés ---
	on := bctx.Params["on"]
	leftKey := bctx.Params["left_on"]
	rightKey := bctx.Params["right_on"]
	if leftKey == "" && on != "" {
		leftKey = on
	}
	if rightKey == "" && on != "" {
		rightKey = on
	}
	if leftKey == "" || rightKey == "" {
		return fmt.Errorf("transform.merge: 'on' ou ('left_on' + 'right_on') requis")
	}

	how := strings.ToLower(bctx.Params["how"])
	if how == "" {
		how = "inner"
	}
	switch how {
	case "inner", "left", "right", "outer":
	default:
		return fmt.Errorf("transform.merge: how='%s' non supporté (inner|left|right|outer)", how)
	}

	leftSuffix := bctx.Params["left_suffix"]
	if leftSuffix == "" {
		leftSuffix = "_x"
	}
	rightSuffix := bctx.Params["right_suffix"]
	if rightSuffix == "" {
		rightSuffix = "_y"
	}

	validate := bctx.Params["validate"]

	// --- Phase build : charger le flux droit en mémoire ---
	rightMap := make(map[string][]contracts.DataRow)
	for row := range bctx.Inputs[1].Ch {
		k := fmt.Sprintf("%v", row[rightKey])
		rightMap[k] = append(rightMap[k], row)
	}

	// --- Validation de cardinalité ---
	if err := validateCardinality(validate, rightMap); err != nil {
		return fmt.Errorf("transform.merge validate: %w", err)
	}

	// Pour outer/right join : tracker les clés droites déjà émises.
	emittedRight := make(map[string]bool)

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

	closeOutputs := func() {
		for _, out := range bctx.Outputs {
			close(out.Ch)
		}
	}

	// --- Phase probe : stream sur le flux gauche ---
	// On doit aussi valider one_to_many / many_to_one sur les clés gauches.
	leftKeyCounts := make(map[string]int)

	for {
		select {
		case <-bctx.Ctx.Done():
			closeOutputs()
			return bctx.Ctx.Err()

		case leftRow, ok := <-bctx.Inputs[0].Ch:
			if !ok {
				// Flux gauche épuisé.
				// Validation many_to_one / one_to_one sur les clés gauches.
				if validate == "many_to_one" || validate == "one_to_one" {
					for k, cnt := range leftKeyCounts {
						if cnt > 1 {
							closeOutputs()
							return fmt.Errorf("transform.merge: validate=%s — clé gauche '%s' apparaît %d fois", validate, k, cnt)
						}
					}
				}
				// outer/right : émettre les lignes droites non matchées.
				if how == "outer" || how == "right" {
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
				closeOutputs()
				return nil
			}

			k := fmt.Sprintf("%v", leftRow[leftKey])
			leftKeyCounts[k]++
			rightRows, matched := rightMap[k]

			if matched {
				emittedRight[k] = true
				for _, rightRow := range rightRows {
					merged := mergeRowsSuffixed(leftRow, rightRow, leftKey, rightKey, leftSuffix, rightSuffix)
					if err := emit(merged); err != nil {
						return err
					}
				}
			} else if how == "left" || how == "outer" {
				if err := emit(leftRow); err != nil {
					return err
				}
			}
		}
	}
}

// mergeRowsSuffixed fusionne deux lignes en appliquant les suffixes Pandas (_x/_y)
// sur les colonnes en collision (hors clés de jointure).
func mergeRowsSuffixed(left, right contracts.DataRow, leftKey, rightKey, leftSuffix, rightSuffix string) contracts.DataRow {
	merged := make(contracts.DataRow, len(left)+len(right))

	// Détecter les collisions (hors clés).
	collisions := make(map[string]bool)
	for k := range right {
		if k == rightKey {
			continue
		}
		if _, exists := left[k]; exists && k != leftKey {
			collisions[k] = true
		}
	}

	// Insérer les colonnes gauches.
	for k, v := range left {
		if collisions[k] {
			merged[k+leftSuffix] = v
		} else {
			merged[k] = v
		}
	}

	// Insérer les colonnes droites.
	for k, v := range right {
		if k == rightKey {
			continue // clé droite déjà représentée par leftKey
		}
		if collisions[k] {
			merged[k+rightSuffix] = v
		} else {
			merged[k] = v
		}
	}
	return merged
}

// validateCardinality vérifie la cardinalité du flux droit (build side).
func validateCardinality(validate string, rightMap map[string][]contracts.DataRow) error {
	switch validate {
	case "one_to_one", "many_to_one":
		for k, rows := range rightMap {
			if len(rows) > 1 {
				return fmt.Errorf("validate=%s — clé droite '%s' apparaît %d fois", validate, k, len(rows))
			}
		}
	case "one_to_many", "many_to_many", "":
		// pas de restriction
	default:
		return fmt.Errorf("validate='%s' non reconnu (one_to_one|one_to_many|many_to_one|many_to_many)", validate)
	}
	return nil
}
