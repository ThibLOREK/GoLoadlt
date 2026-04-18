package analytics

import (
	"fmt"
	"math"
	"strconv"

	"github.com/rinjold/go-etl-studio/internal/etl/blocks"
	"github.com/rinjold/go-etl-studio/internal/etl/contracts"
)

func init() {
	blocks.Register("analytics.field_summary", func() contracts.Block { return &FieldSummary{} })
}

// FieldSummary calcule min, max, mean, nulls, count pour chaque colonne numérique.
// Émet une ligne par colonne analysée.
type FieldSummary struct{}

func (b *FieldSummary) Type() string { return "analytics.field_summary" }

func (b *FieldSummary) Run(bctx *contracts.BlockContext) error {
	type stats struct {
		count    int64
		nulls    int64
		sum      float64
		min, max float64
		isNum    bool
	}

	colStats := make(map[string]*stats)
	var colOrder []string

	for row := range bctx.Inputs[0].Ch {
		for k, v := range row {
			if _, exists := colStats[k]; !exists {
				colStats[k] = &stats{min: math.MaxFloat64, max: -math.MaxFloat64}
				colOrder = append(colOrder, k)
			}
			s := colStats[k]
			s.count++
			if v == nil {
				s.nulls++
				continue
			}
			f, err := strconv.ParseFloat(fmt.Sprintf("%v", v), 64)
			if err == nil {
				s.isNum = true
				s.sum += f
				if f < s.min { s.min = f }
				if f > s.max { s.max = f }
			}
		}
	}

	for _, col := range colOrder {
		s := colStats[col]
		row := contracts.DataRow{
			"column": col,
			"count":  s.count,
			"nulls":  s.nulls,
		}
		if s.isNum && s.count > s.nulls {
			n := float64(s.count - s.nulls)
			row["min"] = s.min
			row["max"] = s.max
			row["mean"] = s.sum / n
		}
		for _, out := range bctx.Outputs { out.Ch <- row }
	}
	for _, out := range bctx.Outputs { close(out.Ch) }
	return nil
}