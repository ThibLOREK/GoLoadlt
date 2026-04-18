package transforms

import (
	"fmt"

	"github.com/rinjold/go-etl-studio/internal/etl/blocks"
	"github.com/rinjold/go-etl-studio/internal/etl/contracts"
)

func init() {
	blocks.Register("transform.append_fields", func() contracts.Block { return &AppendFields{} })
}

// AppendFields ajoute les colonnes du second flux au premier (horizontalement, ligne par ligne).
type AppendFields struct{}

func (b *AppendFields) Type() string { return "transform.append_fields" }

func (b *AppendFields) Run(bctx *contracts.BlockContext) error {
	if len(bctx.Inputs) < 2 {
		return fmt.Errorf("transform.append_fields: 2 ports d'entrée requis")
	}
	for leftRow := range bctx.Inputs[0].Ch {
		rightRow, ok := <-bctx.Inputs[1].Ch
		merged := make(contracts.DataRow, len(leftRow))
		for k, v := range leftRow { merged[k] = v }
		if ok {
			for k, v := range rightRow {
				if _, exists := merged[k]; exists {
					merged["right_"+k] = v
				} else {
					merged[k] = v
				}
			}
		}
		for _, out := range bctx.Outputs { out.Ch <- merged }
	}
	for _, out := range bctx.Outputs { close(out.Ch) }
	return nil
}