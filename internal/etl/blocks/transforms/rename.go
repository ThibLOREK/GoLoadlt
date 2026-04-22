package transforms

import (
	"fmt"
	"strings"

	"github.com/rinjold/go-etl-studio/internal/etl/blocks"
	"github.com/rinjold/go-etl-studio/internal/etl/contracts"
)

func init() {
	blocks.Register("transform.rename", func() contracts.Block { return &Rename{} })
}

// Rename renomme des colonnes — parité complète df.rename().
//
// Paramètres (bctx.Params) :
//   - columns : mapping "ancien:nouveau" séparé par virgule
//               ex: "user_id:id,first_name:prenom,last_name:nom"
//   - errors  : "ignore" (défaut) | "raise"
//               raise = retourne une erreur si une colonne source est absente
type Rename struct{}

func (b *Rename) Type() string { return "transform.rename" }

func (b *Rename) Run(bctx *contracts.BlockContext) error {
	if len(bctx.Inputs) == 0 {
		return fmt.Errorf("transform.rename: aucun port d'entrée")
	}

	colsRaw := bctx.Params["columns"]
	if colsRaw == "" {
		return fmt.Errorf("transform.rename: paramètre 'columns' requis (ex: ancien:nouveau,x:y)")
	}
	errorsMode := strings.ToLower(bctx.Params["errors"])
	if errorsMode == "" {
		errorsMode = "ignore"
	}
	switch errorsMode {
	case "ignore", "raise":
	default:
		return fmt.Errorf("transform.rename: errors='%s' non supporté (ignore|raise)", errorsMode)
	}

	// Parser le mapping ancien→nouveau
	renameMap := make(map[string]string)
	for _, pair := range splitComma(colsRaw) {
		pair = strings.TrimSpace(pair)
		idx := strings.Index(pair, ":")
		if idx < 0 {
			return fmt.Errorf("transform.rename: paire invalide '%s' (attendu 'ancien:nouveau')", pair)
		}
		old := strings.TrimSpace(pair[:idx])
		new_ := strings.TrimSpace(pair[idx+1:])
		if old == "" || new_ == "" {
			return fmt.Errorf("transform.rename: paire invalide '%s'", pair)
		}
		renameMap[old] = new_
	}

	// Si errors=raise, on doit vérifier les noms sur la 1ère ligne
	// (approche streaming : on vérifie au fil de l'eau sur la première ligne reçue)
	firstRow := true

	closeOutputs := func() {
		for _, out := range bctx.Outputs {
			close(out.Ch)
		}
	}

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
			// Validation errors=raise sur la 1ère ligne
			if firstRow && errorsMode == "raise" {
				for old := range renameMap {
					if _, exists := row[old]; !exists {
						closeOutputs()
						return fmt.Errorf("transform.rename: colonne '%s' absente du flux (errors=raise)", old)
					}
				}
				firstRow = false
			}
			firstRow = false

			// Appliquer le renommage
			out := make(contracts.DataRow, len(row))
			for col, val := range row {
				if newName, ok := renameMap[col]; ok {
					out[newName] = val
				} else {
					out[col] = val
				}
			}
			for _, outPort := range bctx.Outputs {
				select {
				case outPort.Ch <- out:
				case <-bctx.Ctx.Done():
					closeOutputs()
					return bctx.Ctx.Err()
				}
			}
		}
	}
}
