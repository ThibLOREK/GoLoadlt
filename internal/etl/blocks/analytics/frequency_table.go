package analytics

import (
	"fmt"
	"sort"

	"github.com/rinjold/go-etl-studio/internal/etl/blocks"
	"github.com/rinjold/go-etl-studio/internal/etl/contracts"
)

func init() {
	blocks.Register("analytics.frequency_table", func() contracts.Block { return &FrequencyTable{} })
}

// FrequencyTable calcule la fréquence de chaque valeur distincte d'une colonne.
// Émet : value, count, percent
type FrequencyTable struct{}

func (b *FrequencyTable) Type() string { return "analytics.frequency_table" }

func (b *FrequencyTable) Run(bctx *contracts.BlockContext) error {
	col := bctx.Params["column"]
	if col == "" {
		return fmt.Errorf("analytics.frequency_table: 'column' obligatoire")
	}

	freq := make(map[string]int64)
	var total int64
	for row := range bctx.Inputs[0].Ch {
		key := fmt.Sprintf("%v", row[col])
		freq[key]++
		total++
	}

	type kv struct {
		k string
		v int64
	}
	var sorted []kv
	for k, v := range freq {
		sorted = append(sorted, kv{k, v})
	}
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].v > sorted[j].v })

	for _, item := range sorted {
		pct := 0.0
		if total > 0 {
			pct = float64(item.v) / float64(total) * 100
		}
		outRow := contracts.DataRow{
			"value":   item.k,
			"count":   item.v,
			"percent": fmt.Sprintf("%.2f", pct),
		}
		for _, out := range bctx.Outputs {
			out.Ch <- outRow
		}
	}
	for _, out := range bctx.Outputs {
		close(out.Ch)
	}
	return nil
}
