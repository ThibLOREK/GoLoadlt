package transforms

import (
	"fmt"
	"strings"

	"github.com/rinjold/go-etl-studio/internal/etl/blocks"
	"github.com/rinjold/go-etl-studio/internal/etl/contracts"
)

func init() {
	blocks.Register("transform.fillna", func() contracts.Block { return &Fillna{} })
}

// Fillna remplace les valeurs nulles/vides — parité complète df.fillna().
//
// Paramètres (bctx.Params) :
//   - value   : valeur de remplacement scalaire (ex: "0", "inconnu")
//   - columns : colonnes à traiter, séparées par virgule (vide = toutes les colonnes)
//   - method  : "" (valeur scalaire) | "ffill" (propagation avant) | "bfill" (propagation arrière)
//   - limit   : nombre max de remplissages consécutifs ("" = illimité), ex: "3"
//
// Stratégie streaming :
//   - Scalaire  : stream pur, O(1) mémoire.
//   - ffill     : stream avec buffer de la dernière valeur non-nulle par colonne.
//   - bfill     : chargement complet en mémoire (nécessaire pour regarder en avant).
type Fillna struct{}

func (b *Fillna) Type() string { return "transform.fillna" }

func isNull(v any) bool {
	if v == nil {
		return true
	}
	s, ok := v.(string)
	return ok && (s == "" || strings.EqualFold(s, "null") || strings.EqualFold(s, "<nil>") || strings.EqualFold(s, "nan"))
}

func (b *Fillna) Run(bctx *contracts.BlockContext) error {
	if len(bctx.Inputs) == 0 {
		return fmt.Errorf("transform.fillna: aucun port d'entrée")
	}

	value := bctx.Params["value"]
	method := strings.ToLower(bctx.Params["method"])
	colsCSV := bctx.Params["columns"]
	limitStr := bctx.Params["limit"]

	// Validation
	if method == "" && value == "" {
		return fmt.Errorf("transform.fillna: 'value' ou 'method' (ffill|bfill) requis")
	}
	switch method {
	case "", "ffill", "bfill":
	default:
		return fmt.Errorf("transform.fillna: method='%s' non supporté (ffill|bfill)", method)
	}

	limit := -1 // -1 = illimité
	if limitStr != "" {
		fmt.Sscanf(limitStr, "%d", &limit)
	}

	// Colonnes cibles (nil = toutes)
	var targetCols []string
	if colsCSV != "" {
		targetCols = splitComma(colsCSV)
	}

	closeOutputs := func() {
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

	switch method {

	case "": // --- Remplacement scalaire (stream pur) ---
		for {
			select {
			case <-bctx.Ctx.Done():
				closeOutputs()
				return bctx.Ctx.Err()
			case row, ok := <-bctx.Inputs[0].Ch:
				if !ok {
					closeOutputs()
					return nil
				}
				out := copyRow(row)
				for col, v := range out {
					if !matchesTarget(col, targetCols) {
						continue
					}
					if isNull(v) {
						out[col] = value
					}
				}
				if err := emit(out); err != nil {
					return err
				}
			}
		}

	case "ffill": // --- Forward fill (stream avec last-seen) ---
		lastSeen := make(map[string]any)
		consecCount := make(map[string]int)
		for {
			select {
			case <-bctx.Ctx.Done():
				closeOutputs()
				return bctx.Ctx.Err()
			case row, ok := <-bctx.Inputs[0].Ch:
				if !ok {
					closeOutputs()
					return nil
				}
				out := copyRow(row)
				for col, v := range out {
					if !matchesTarget(col, targetCols) {
						continue
					}
					if isNull(v) {
						if prev, exists := lastSeen[col]; exists {
							if limit < 0 || consecCount[col] < limit {
								out[col] = prev
								consecCount[col]++
							}
						}
					} else {
						lastSeen[col] = v
						consecCount[col] = 0
					}
				}
				if err := emit(out); err != nil {
					return err
				}
			}
		}

	case "bfill": // --- Backward fill (chargement complet en mémoire) ---
		var rows []contracts.DataRow
		for row := range bctx.Inputs[0].Ch {
			rows = append(rows, copyRow(row))
		}
		// Passe arrière par colonne
		for col := range rows[len(rows)-1] {
			if !matchesTarget(col, targetCols) {
				continue
			}
			var next any
			count := 0
			for i := len(rows) - 1; i >= 0; i-- {
				v := rows[i][col]
				if !isNull(v) {
					next = v
					count = 0
				} else if next != nil {
					if limit < 0 || count < limit {
						rows[i][col] = next
						count++
					}
				}
			}
		}
		for _, row := range rows {
			if err := emit(row); err != nil {
				closeOutputs()
				return err
			}
		}
		closeOutputs()
	}
	return nil
}

// matchesTarget retourne true si la colonne doit être traitée.
func matchesTarget(col string, targets []string) bool {
	if len(targets) == 0 {
		return true // toutes les colonnes
	}
	for _, t := range targets {
		if t == col {
			return true
		}
	}
	return false
}

// copyRow duplique une DataRow pour éviter les mutations.
func copyRow(row contracts.DataRow) contracts.DataRow {
	out := make(contracts.DataRow, len(row))
	for k, v := range row {
		out[k] = v
	}
	return out
}
