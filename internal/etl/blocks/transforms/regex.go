package transforms

import (
	"fmt"
	"regexp"

	"github.com/rinjold/go-etl-studio/internal/etl/blocks"
	"github.com/rinjold/go-etl-studio/internal/etl/contracts"
)

func init() {
	blocks.Register("transform.regex", func() contracts.Block { return &RegEx{} })
}

// RegEx applique une expression régulière sur une colonne.
// Paramètres :
//   column   : colonne source
//   pattern  : expression régulière
//   mode     : extract (1er groupe capturant) | replace | match (filtre les lignes qui matchent)
//   replace  : valeur de remplacement (si mode=replace)
//   output   : colonne de sortie (si mode=extract, défaut: column+"_extracted")
type RegEx struct{}

func (b *RegEx) Type() string { return "transform.regex" }

func (b *RegEx) Run(bctx *contracts.BlockContext) error {
	col := bctx.Params["column"]
	pattern := bctx.Params["pattern"]
	mode := bctx.Params["mode"]
	if mode == "" { mode = "extract" }
	if col == "" || pattern == "" {
		return fmt.Errorf("transform.regex: 'column' et 'pattern' obligatoires")
	}

	re, err := regexp.Compile(pattern)
	if err != nil {
		return fmt.Errorf("transform.regex: pattern invalide '%s': %w", pattern, err)
	}

	outputCol := bctx.Params["output"]
	if outputCol == "" { outputCol = col + "_extracted" }
	replaceWith := bctx.Params["replace"]

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
			str := fmt.Sprintf("%v", row[col])
			newRow := make(contracts.DataRow, len(row)+1)
			for k, v := range row { newRow[k] = v }

			switch mode {
			case "extract":
				m := re.FindStringSubmatch(str)
				if len(m) > 1 { newRow[outputCol] = m[1] } else { newRow[outputCol] = "" }
				for _, out := range bctx.Outputs { out.Ch <- newRow }
			case "replace":
				newRow[col] = re.ReplaceAllString(str, replaceWith)
				for _, out := range bctx.Outputs { out.Ch <- newRow }
			case "match":
				if re.MatchString(str) {
					for _, out := range bctx.Outputs { out.Ch <- newRow }
				}
			}
		}
	}
}