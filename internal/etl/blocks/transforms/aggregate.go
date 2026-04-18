package transforms

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/rinjold/go-etl-studio/internal/etl/blocks"
	"github.com/rinjold/go-etl-studio/internal/etl/contracts"
)

func init() {
	blocks.Register("transform.aggregate", func() contracts.Block { return &Aggregate{} })
}

// Aggregate effectue SUM/COUNT/AVG/MIN/MAX par groupe.
// Params:
//   - groupBy      : colonnes de groupe séparées par virgule (ex: "region,category")
//   - aggregations : liste CSV de "func(colonne)" (ex: "SUM(amount),COUNT(id),AVG(price)")
type Aggregate struct{}

func (b *Aggregate) Type() string { return "transform.aggregate" }

type aggSpec struct {
	func_ string
	col   string
	alias string
}

func parseAggSpecs(raw string) []aggSpec {
	var specs []aggSpec
	for _, part := range splitComma(raw) {
		part = strings.TrimSpace(part)
		open := strings.Index(part, "(")
		close_ := strings.Index(part, ")")
		if open < 0 || close_ < 0 {
			continue
		}
		fn := strings.ToUpper(strings.TrimSpace(part[:open]))
		col := strings.TrimSpace(part[open+1 : close_])
		specs = append(specs, aggSpec{func_: fn, col: col, alias: fn + "_" + col})
	}
	return specs
}

func (b *Aggregate) Run(bctx *contracts.BlockContext) error {
	groupByCSV := bctx.Params["groupBy"]
	aggsRaw := bctx.Params["aggregations"]
	if groupByCSV == "" || aggsRaw == "" {
		return fmt.Errorf("transform.aggregate: paramètres 'groupBy' et 'aggregations' requis")
	}
	if len(bctx.Inputs) == 0 {
		return fmt.Errorf("transform.aggregate: aucun port d'entrée")
	}
	groupByCols := splitComma(groupByCSV)
	specs := parseAggSpecs(aggsRaw)

	type accumulator map[string][]float64
	groups := make(map[string]accumulator)
	groupKeys := []string{}
	seen := map[string]bool{}

	for row := range bctx.Inputs[0].Ch {
		parts := make([]string, len(groupByCols))
		for i, c := range groupByCols {
			parts[i] = fmt.Sprintf("%v", row[c])
		}
		gk := strings.Join(parts, "||")
		if !seen[gk] {
			seen[gk] = true
			groupKeys = append(groupKeys, gk)
			groups[gk] = make(accumulator)
		}
		for _, spec := range specs {
			v, _ := strconv.ParseFloat(fmt.Sprintf("%v", row[spec.col]), 64)
			groups[gk][spec.alias] = append(groups[gk][spec.alias], v)
		}
	}

	sort.Strings(groupKeys)

	for _, gk := range groupKeys {
		parts := strings.Split(gk, "||")
		row := make(contracts.DataRow)
		for i, c := range groupByCols {
			row[c] = parts[i]
		}
		acc := groups[gk]
		for _, spec := range specs {
			row[spec.alias] = aggregate(spec.func_, acc[spec.alias])
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
