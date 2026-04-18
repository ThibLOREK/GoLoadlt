package transforms

import (
	"strconv"

	"github.com/rinjold/go-etl-studio/internal/etl/blocks"
	"github.com/rinjold/go-etl-studio/internal/etl/contracts"
)

func init() {
	blocks.Register("transform.auto_field", func() contracts.Block { return &AutoField{} })
}

// AutoField détecte et convertit automatiquement les types de colonnes.
// Ordre de détection : int64 → float64 → bool → string
type AutoField struct{}

func (b *AutoField) Type() string { return "transform.auto_field" }

func (b *AutoField) Run(bctx *contracts.BlockContext) error {
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
			newRow := make(contracts.DataRow, len(row))
			for k, v := range row {
				str, isStr := v.(string)
				if !isStr {
					newRow[k] = v
					continue
				}
				if i, err := strconv.ParseInt(str, 10, 64); err == nil {
					newRow[k] = i
				} else if f, err := strconv.ParseFloat(str, 64); err == nil {
					newRow[k] = f
				} else if b2, err := strconv.ParseBool(str); err == nil {
					newRow[k] = b2
				} else {
					newRow[k] = str
				}
			}
			for _, out := range bctx.Outputs { out.Ch <- newRow }
		}
	}
}