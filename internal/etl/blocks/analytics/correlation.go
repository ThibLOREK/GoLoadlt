package analytics

import (
	"fmt"
	"math"
	"strconv"

	"github.com/rinjold/go-etl-studio/internal/etl/blocks"
	"github.com/rinjold/go-etl-studio/internal/etl/contracts"
)

func init() {
	blocks.Register("analytics.correlation", func() contracts.Block { return &Correlation{} })
}

// Correlation calcule la corrélation de Pearson entre toutes les paires de colonnes numériques.
// Émet : col1, col2, pearson
type Correlation struct{}

func (b *Correlation) Type() string { return "analytics.correlation" }

func (b *Correlation) Run(bctx *contracts.BlockContext) error {
	var rows []contracts.DataRow
	var numCols []string

	for row := range bctx.Inputs[0].Ch {
		if numCols == nil {
			for k, v := range row {
				if _, err := strconv.ParseFloat(fmt.Sprintf("%v", v), 64); err == nil {
					numCols = append(numCols, k)
				}
			}
		}
		rows = append(rows, row)
	}

	extract := func(col string) []float64 {
		vals := make([]float64, 0, len(rows))
		for _, r := range rows {
			if f, err := strconv.ParseFloat(fmt.Sprintf("%v", r[col]), 64); err == nil {
				vals = append(vals, f)
			}
		}
		return vals
	}

	pearson := func(xs, ys []float64) float64 {
		n := float64(len(xs))
		if n == 0 { return 0 }
		var sumX, sumY, sumXY, sumX2, sumY2 float64
		for i := range xs {
			sumX += xs[i]; sumY += ys[i]
			sumXY += xs[i] * ys[i]
			sumX2 += xs[i] * xs[i]
			sumY2 += ys[i] * ys[i]
		}
		num := n*sumXY - sumX*sumY
		den := math.Sqrt((n*sumX2 - sumX*sumX) * (n*sumY2 - sumY*sumY))
		if den == 0 { return 0 }
		return num / den
	}

	for i := 0; i < len(numCols); i++ {
		for j := i + 1; j < len(numCols); j++ {
			xs := extract(numCols[i])
			ys := extract(numCols[j])
			r := pearson(xs, ys)
			outRow := contracts.DataRow{
				"col1":    numCols[i],
				"col2":    numCols[j],
				"pearson": fmt.Sprintf("%.4f", r),
			}
			for _, out := range bctx.Outputs { out.Ch <- outRow }
		}
	}
	for _, out := range bctx.Outputs { close(out.Ch) }
	return nil
}