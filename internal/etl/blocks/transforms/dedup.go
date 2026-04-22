package transforms

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/rinjold/go-etl-studio/internal/etl/blocks"
	"github.com/rinjold/go-etl-studio/internal/etl/contracts"
)

func init() {
	blocks.Register("transform.dedup", func() contracts.Block { return &Dedup{} })
}

// Dedup supprime les doublons selon une ou plusieurs colonnes clés.
//
// Paramètres (bctx.Params) — trois formats acceptés, priorité dans cet ordre :
//
//  1. Indexé  : key_0="col_1", key_1="col_2", …  (format natif UI liste déroulante multi-entrées)
//  2. JSON    : keys="[\"col_1\",\"col_2\"]"       (array JSON)
//  3. CSV     : keys="col_1,col_2"               (rétrocompatibilité)
//
// Si aucun paramètre n'est fourni, le bloc dédoublonne sur TOUTES les colonnes
// détectées sur la première ligne reçue.
type Dedup struct{}

func (b *Dedup) Type() string { return "transform.dedup" }

func (b *Dedup) Run(bctx *contracts.BlockContext) error {
	if len(bctx.Inputs) == 0 {
		return fmt.Errorf("transform.dedup: aucun port d'entrée")
	}

	// --- Résolution des clés ---
	keys := resolveKeys(bctx)

	// Si les clés sont encore inconnues (aucun param), on les découvrira
	// depuis le schéma déclaré du port d'entrée.
	if len(keys) == 0 && len(bctx.Inputs[0].Schema.Columns) > 0 {
		for _, c := range bctx.Inputs[0].Schema.Columns {
			keys = append(keys, c.Name)
		}
	}

	// schemaDiscovered : true quand les clés seront extraites de la 1ère ligne
	schemaDiscovered := len(keys) == 0

	closeOutputs := func() {
		for _, out := range bctx.Outputs {
			close(out.Ch)
		}
	}

	seen := make(map[string]bool)
	in := bctx.Inputs[0]

	for {
		select {
		case <-bctx.Ctx.Done():
			closeOutputs()
			return bctx.Ctx.Err()
		case row, ok := <-in.Ch:
			if !ok {
				closeOutputs()
				return nil
			}

			// Découverte du schéma sur la 1ère ligne reçue
			if schemaDiscovered {
				for col := range row {
					keys = append(keys, col)
				}
				schemaDiscovered = false
			}

			// Clé composite
			parts := make([]string, len(keys))
			for i, k := range keys {
				parts[i] = fmt.Sprintf("%v", row[k])
			}
			composite := strings.Join(parts, "\x00")

			if seen[composite] {
				continue
			}
			seen[composite] = true

			for _, out := range bctx.Outputs {
				select {
				case out.Ch <- row:
				case <-bctx.Ctx.Done():
					closeOutputs()
					return bctx.Ctx.Err()
				}
			}
		}
	}
}

// resolveKeys extrait les colonnes clés depuis bctx.Params.
// Priorité : indexé (key_0…) > JSON array > CSV.
func resolveKeys(bctx *contracts.BlockContext) []string {
	// 1. Format indexé : key_0, key_1, …
	var indexed []string
	for i := 0; ; i++ {
		v := strings.TrimSpace(bctx.Params[fmt.Sprintf("key_%d", i)])
		if v == "" {
			break
		}
		indexed = append(indexed, v)
	}
	if len(indexed) > 0 {
		return indexed
	}

	// 2. Format JSON array : keys="[\"col_1\",\"col_2\"]"
	raw := strings.TrimSpace(bctx.Params["keys"])
	if strings.HasPrefix(raw, "[") {
		var arr []string
		if err := json.Unmarshal([]byte(raw), &arr); err == nil && len(arr) > 0 {
			return arr
		}
	}

	// 3. Format CSV : keys="col_1,col_2"
	if raw != "" {
		return splitComma(raw)
	}

	return nil
}
