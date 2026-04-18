package transforms

import (
	"fmt"
	"sort"
	"strconv"

	"github.com/rinjold/go-etl-studio/internal/etl/blocks"
	"github.com/rinjold/go-etl-studio/internal/etl/contracts"
)

func init() {
	blocks.Register("transform.pivot", func() contracts.Block { return &Pivot{} })
}

// Pivot groupe les lignes par groupBy et pivote la colonne valueColumn en colonnes distinctes.
// Params:
//   - groupBy     : colonne de regroupement (ex: "region")
//   - pivotColumn : colonne dont les valeurs deviennent des colonnes (ex: "product")
//   - valueColumn : colonne à agréger (ex: "amount")
//   - aggregation : SUM | COUNT | AVG | MIN | MAX (défaut: SUM)
type Pivot struct{}

func (b *Pivot) Type() string { return "transform.pivot" }

func (b *Pivot) Run(bctx *contracts.BlockContext) error {
	groupBy := bctx.Params["groupBy"]
	pivotCol := bctx.Params["pivotColumn"]
	valueCol := bctx.Params["valueColumn"]
	agg := bctx.Params["aggregation"]
	if agg == "" {
		agg = "SUM"
	}
	if groupBy == "" || pivotCol == "" || valueCol == "" {
		return fmt.Errorf("transform.pivot: paramètres 'groupBy', 'pivotColumn', 'valueColumn' requis")
	}
	if len(bctx.Inputs) == 0 {
		return fmt.Errorf("transform.pivot: aucun port d'entrée")
	}

	// Collecter toutes les lignes en mémoire.
	type key struct{ group, pivot string }
	aggMap := make(map[key][]float64)
	var groups, pivots []string
	groupSet := map[string]bool{}
	pivotSet := map[string]bool{}

	for row := range bctx.Inputs[0].Ch {
		g := fmt.Sprintf("%v", row[groupBy])
		p := fmt.Sprintf("%v", row[pivotCol])
		v, _ := strconv.ParseFloat(fmt.Sprintf("%v", row[valueCol]), 64)
		aggMap[key{g, p}] = append(aggMap[key{g, p}], v)
		if !groupSet[g] {
			groupSet[g] = true
			groups = append(groups, g)
		}
		if !pivotSet[p] {
			pivotSet[p] = true
			pivots = append(pivots, p)
		}
	}
	sort.Strings(groups)
	sort.Strings(pivots)

	for _, g := range groups {
		row := contracts.DataRow{groupBy: g}
		for _, p := range pivots {
			vals := aggMap[key{g, p}]
			row[p] = aggregate(agg, vals)
		}
		for _, out := range bctx.Outputs {
			select {
			case out.Ch <- row:
			case <-bctx.Ctx.Done():
				return bctx.Ctx.Err()
			}
		}
	}
	for _, out := range bctx.Outputs {
		close(out.Ch)
	}
	return nil
}

func aggregate(agg string, vals []float64) float64 {
	if len(vals) == 0 {
		return 0
	}
	switch agg {
	case "COUNT":
		return float64(len(vals))
	case "AVG":
		var s float64
		for _, v := range vals {
			s += v
		}
		return s / float64(len(vals))
	case "MIN":
		m := vals[0]
		for _, v := range vals {
			if v < m {
				m = v
			}
		}
		return m
	case "MAX":
		m := vals[0]
		for _, v := range vals {
			if v > m {
				m = v
			}
		}
		return m
	default: // SUM
		var s float64
		for _, v := range vals {
			s += v
		}
		return s
	}
}
