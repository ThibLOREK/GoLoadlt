package transforms

import (
	"fmt"
	"strings"

	"github.com/rinjold/go-etl-studio/internal/etl/blocks"
	"github.com/rinjold/go-etl-studio/internal/etl/contracts"
)

func init() {
	blocks.Register("transform.select", func() contracts.Block { return &Select{} })
}

// Select sélectionne et/ou renomme des colonnes.
// Paramètre 'columns' : liste séparée par virgules.
// Syntaxe : "col1, col2 as alias2, col3"
type Select struct{}

func (b *Select) Type() string { return "transform.select" }

func (b *Select) Run(bctx *contracts.BlockContext) error {
	colsParam := bctx.Params["columns"]
	if colsParam == "" {
		return fmt.Errorf("transform.select: paramètre 'columns' manquant")
	}

	type colMapping struct {
		src   string
		dst   string
	}
	var mappings []colMapping
	for _, part := range strings.Split(colsParam, ",") {
		part = strings.TrimSpace(part)
		if asIdx := strings.Index(strings.ToLower(part), " as "); asIdx >= 0 {
			src := strings.TrimSpace(part[:asIdx])
			dst := strings.TrimSpace(part[asIdx+4:])
			mappings = append(mappings, colMapping{src, dst})
		} else {
			mappings = append(mappings, colMapping{part, part})
		}
	}

	if len(bctx.Inputs) == 0 {
		return fmt.Errorf("transform.select: aucun port d'entrée")
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
			newRow := make(contracts.DataRow, len(mappings))
			for _, m := range mappings {
				newRow[m.dst] = row[m.src]
			}
			for _, out := range bctx.Outputs { out.Ch <- newRow }
		}
	}
}
