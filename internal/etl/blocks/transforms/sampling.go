package transforms

import (
	"fmt"
	"math/rand"
	"strconv"

	"github.com/rinjold/go-etl-studio/internal/etl/blocks"
	"github.com/rinjold/go-etl-studio/internal/etl/contracts"
)

func init() {
	blocks.Register("transform.sampling", func() contracts.Block { return &Sampling{} })
}

// Sampling échantillonne le flux.
// Paramètres :
//   mode    : first (N premières lignes) | percent (% aléatoire) | every (1 ligne sur N)
//   value   : N ou pourcentage (ex: "100", "10.5", "5")
type Sampling struct{}

func (b *Sampling) Type() string { return "transform.sampling" }

func (b *Sampling) Run(bctx *contracts.BlockContext) error {
	mode := bctx.Params["mode"]
	valueStr := bctx.Params["value"]
	if mode == "" { mode = "first" }
	value, _ := strconv.ParseFloat(valueStr, 64)
	if value <= 0 { return fmt.Errorf("transform.sampling: 'value' doit être > 0") }

	in := bctx.Inputs[0]
	count := 0

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
			count++
			emit := false
			switch mode {
			case "first":
				emit = float64(count) <= value
			case "percent":
				emit = rand.Float64()*100 < value
			case "every":
				emit = count%int(value) == 1
			}
			if emit {
				for _, out := range bctx.Outputs { out.Ch <- row }
			}
		}
	}
}