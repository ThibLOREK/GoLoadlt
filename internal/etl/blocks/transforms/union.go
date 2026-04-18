package transforms

import (
	"fmt"
	"sync"

	"github.com/rinjold/go-etl-studio/internal/etl/blocks"
	"github.com/rinjold/go-etl-studio/internal/etl/contracts"
)

func init() {
	blocks.Register("transform.union", func() contracts.Block { return &Union{} })
}

// Union fusionne N flux en un seul (comme UNION ALL en SQL).
type Union struct{}

func (b *Union) Type() string { return "transform.union" }

func (b *Union) Run(bctx *contracts.BlockContext) error {
	if len(bctx.Inputs) < 2 {
		return fmt.Errorf("transform.union: au moins 2 ports d'entrée requis")
	}
	var wg sync.WaitGroup
	for _, in := range bctx.Inputs {
		in := in
		wg.Add(1)
		go func() {
			defer wg.Done()
			for row := range in.Ch {
				for _, out := range bctx.Outputs {
					out.Ch <- row
				}
			}
		}()
	}
	wg.Wait()
	for _, out := range bctx.Outputs { close(out.Ch) }
	return nil
}