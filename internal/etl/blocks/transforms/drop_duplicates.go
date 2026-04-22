package transforms

import (
	"fmt"
	"strings"

	"github.com/rinjold/go-etl-studio/internal/etl/blocks"
	"github.com/rinjold/go-etl-studio/internal/etl/contracts"
)

func init() {
	blocks.Register("transform.drop_duplicates", func() contracts.Block { return &DropDuplicates{} })
}

// DropDuplicates supprime les doublons — parité complète df.drop_duplicates().
//
// Paramètres (bctx.Params) :
//   - subset       : colonnes clés séparées par virgule (vide = toutes les colonnes)
//   - keep         : "first" (défaut) | "last" | "false" (supprimer tous les doublons)
//   - ignore_index : "false" (défaut) | "true" (ajouter colonne "_index" auto-incrémentée)
//
// Stratégie :
//   - keep=first  : stream pur O(n) mémoire sur les clés vues
//   - keep=last   : chargement complet + dedup arrière
//   - keep=false  : deux passes — comptage puis émission des uniques
type DropDuplicates struct{}

func (b *DropDuplicates) Type() string { return "transform.drop_duplicates" }

func (b *DropDuplicates) Run(bctx *contracts.BlockContext) error {
	if len(bctx.Inputs) == 0 {
		return fmt.Errorf("transform.drop_duplicates: aucun port d'entrée")
	}

	subsetCSV := bctx.Params["subset"]
	keep := strings.ToLower(bctx.Params["keep"])
	if keep == "" {
		keep = "first"
	}
	switch keep {
	case "first", "last", "false":
	default:
		return fmt.Errorf("transform.drop_duplicates: keep='%s' non supporté (first|last|false)", keep)
	}
	ignoreIndex := bctx.Params["ignore_index"] == "true"

	var subsetCols []string
	if subsetCSV != "" {
		subsetCols = splitComma(subsetCSV)
	}

	// Fonction de clé composite
	rowKey := func(row contracts.DataRow) string {
		if len(subsetCols) == 0 {
			// Toutes les colonnes : tri des clés pour stabilité
			keys := make([]string, 0, len(row))
			for k := range row {
				keys = append(keys, k)
			}
			// Tri pour ordre stable
			for i := 1; i < len(keys); i++ {
				for j := i; j > 0 && keys[j] < keys[j-1]; j-- {
					keys[j], keys[j-1] = keys[j-1], keys[j]
				}
			}
			parts := make([]string, len(keys))
			for i, k := range keys {
				parts[i] = k + "=" + fmt.Sprintf("%v", row[k])
			}
			return strings.Join(parts, "\x00")
		}
		parts := make([]string, len(subsetCols))
		for i, c := range subsetCols {
			parts[i] = fmt.Sprintf("%v", row[c])
		}
		return strings.Join(parts, "\x00")
	}

	closeOutputs := func() {
		for _, out := range bctx.Outputs {
			close(out.Ch)
		}
	}

	emit := func(row contracts.DataRow, idx int) error {
		if ignoreIndex {
			row["_index"] = idx
		}
		for _, out := range bctx.Outputs {
			select {
			case out.Ch <- row:
			case <-bctx.Ctx.Done():
				return bctx.Ctx.Err()
			}
		}
		return nil
	}

	switch keep {

	case "first": // --- Stream pur ---
		seen := make(map[string]bool)
		idx := 0
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
				k := rowKey(row)
				if seen[k] {
					continue
				}
				seen[k] = true
				if err := emit(row, idx); err != nil {
					return err
				}
				idx++
			}
		}

	case "last": // --- Chargement complet, dédup arrière ---
		var rows []contracts.DataRow
		for row := range bctx.Inputs[0].Ch {
			rows = append(rows, row)
		}
		seen := make(map[string]bool)
		// Passe arrière pour marquer les doublons
		keepIdx := make([]bool, len(rows))
		for i := len(rows) - 1; i >= 0; i-- {
			k := rowKey(rows[i])
			if !seen[k] {
				seen[k] = true
				keepIdx[i] = true
			}
		}
		idx := 0
		for i, row := range rows {
			if keepIdx[i] {
				if err := emit(row, idx); err != nil {
					closeOutputs()
					return err
				}
				idx++
			}
		}
		closeOutputs()

	case "false": // --- Deux passes : compter puis émettre les uniques ---
		var rows []contracts.DataRow
		counts := make(map[string]int)
		for row := range bctx.Inputs[0].Ch {
			rows = append(rows, row)
			counts[rowKey(row)]++
		}
		idx := 0
		for _, row := range rows {
			if counts[rowKey(row)] == 1 {
				if err := emit(row, idx); err != nil {
					closeOutputs()
					return err
				}
				idx++
			}
		}
		closeOutputs()
	}
	return nil
}
