package sources

import (
	"encoding/json"
	"fmt"

	"github.com/rinjold/go-etl-studio/internal/etl/blocks"
	"github.com/rinjold/go-etl-studio/internal/etl/contracts"
)

func init() {
	blocks.Register("source.data_grid", func() contracts.Block { return &DataGrid{} })
}

// DataGrid émet des lignes définies en dur dans la configuration du pipeline.
// Équivalent du step "Data Grid" de Pentaho PDI.
//
// Paramètres (bctx.Params) :
//   - columns : noms des colonnes séparés par virgule
//               ex: "id,name,amount"
//   - rows    : tableau JSON de tableaux de valeurs (ordre = même que columns)
//               ex: [["1","Alice","100"],["2","Bob","200"]]
//
// Exemple complet :
//   {
//     "type": "source.data_grid",
//     "params": {
//       "columns": "id,name,amount",
//       "rows":    "[[\"1\",\"Alice\",\"100\"],[\"2\",\"Bob\",\"200\"]]"
//     }
//   }
type DataGrid struct{}

func (b *DataGrid) Type() string { return "source.data_grid" }

func (b *DataGrid) Run(bctx *contracts.BlockContext) error {
	if len(bctx.Outputs) == 0 {
		return fmt.Errorf("source.data_grid: aucun port de sortie")
	}

	colsCSV := bctx.Params["columns"]
	rowsJSON := bctx.Params["rows"]

	if colsCSV == "" {
		return fmt.Errorf("source.data_grid: paramètre 'columns' requis")
	}
	if rowsJSON == "" {
		return fmt.Errorf("source.data_grid: paramètre 'rows' requis (tableau JSON)")
	}

	// Parser les colonnes
	cols := splitComma(colsCSV)
	if len(cols) == 0 {
		return fmt.Errorf("source.data_grid: 'columns' est vide")
	}

	// Parser les lignes JSON : [["v1","v2",...], ...]
	var rawRows [][]any
	if err := json.Unmarshal([]byte(rowsJSON), &rawRows); err != nil {
		// Essayer aussi [][]string pour la compatibilité
		var strRows [][]string
		if err2 := json.Unmarshal([]byte(rowsJSON), &strRows); err2 != nil {
			return fmt.Errorf("source.data_grid: 'rows' JSON invalide: %w", err)
		}
		rawRows = make([][]any, len(strRows))
		for i, sr := range strRows {
			rawRows[i] = make([]any, len(sr))
			for j, v := range sr {
				rawRows[i][j] = v
			}
		}
	}

	closeOutputs := func() {
		for _, out := range bctx.Outputs {
			close(out.Ch)
		}
	}

	for rowIdx, rawRow := range rawRows {
		// Validation longueur
		if len(rawRow) != len(cols) {
			closeOutputs()
			return fmt.Errorf(
				"source.data_grid: ligne %d a %d valeur(s), attendu %d (colonnes: %v)",
				rowIdx, len(rawRow), len(cols), cols,
			)
		}

		// Construire la DataRow
		row := make(contracts.DataRow, len(cols))
		for i, col := range cols {
			row[col] = rawRow[i]
		}

		// Émettre vers toutes les sorties
		for _, out := range bctx.Outputs {
			select {
			case out.Ch <- row:
			case <-bctx.Ctx.Done():
				closeOutputs()
				return bctx.Ctx.Err()
			}
		}
	}

	closeOutputs()
	return nil
}

// splitComma est dupliquée ici pour éviter une dépendance circulaire avec transforms.
// Si le package expose déjà une version partagée, supprimer cette copie.
func splitComma(s string) []string {
	var out []string
	start := 0
	for i := 0; i <= len(s); i++ {
		if i == len(s) || s[i] == ',' {
			v := trimSpace(s[start:i])
			if v != "" {
				out = append(out, v)
			}
			start = i + 1
		}
	}
	return out
}

func trimSpace(s string) string {
	for len(s) > 0 && (s[0] == ' ' || s[0] == '\t') {
		s = s[1:]
	}
	for len(s) > 0 && (s[len(s)-1] == ' ' || s[len(s)-1] == '\t') {
		s = s[:len(s)-1]
	}
	return s
}
