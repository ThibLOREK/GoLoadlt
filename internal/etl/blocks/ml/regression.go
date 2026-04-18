package ml

import (
	"fmt"
	"math"
	"strconv"

	"github.com/rinjold/go-etl-studio/internal/etl/blocks"
	"github.com/rinjold/go-etl-studio/internal/etl/contracts"
)

func init() {
	blocks.Register("ml.regression", func() contracts.Block { return &LinearRegression{} })
}

// LinearRegression effectue une régression linéaire simple (OLS).
// Paramètres :
//
//	feature  : colonne X (variable explicative)
//	target   : colonne Y (variable cible)
//	output   : colonne de sortie pour la prédiction (défaut: "predicted")
//
// Émet chaque ligne originale enrichie de "predicted", "slope", "intercept", "r2"
type LinearRegression struct{}

func (b *LinearRegression) Type() string { return "ml.regression" }

func (b *LinearRegression) Run(bctx *contracts.BlockContext) error {
	feature := bctx.Params["feature"]
	target := bctx.Params["target"]
	outCol := bctx.Params["output"]
	if feature == "" || target == "" {
		return fmt.Errorf("ml.regression: 'feature' et 'target' obligatoires")
	}
	if outCol == "" {
		outCol = "predicted"
	}

	var rows []contracts.DataRow
	var xs, ys []float64

	for row := range bctx.Inputs[0].Ch {
		rows = append(rows, row)
		x, errX := strconv.ParseFloat(fmt.Sprintf("%v", row[feature]), 64)
		y, errY := strconv.ParseFloat(fmt.Sprintf("%v", row[target]), 64)
		if errX == nil && errY == nil {
			xs = append(xs, x)
			ys = append(ys, y)
		}
	}

	n := float64(len(xs))
	if n == 0 {
		for _, out := range bctx.Outputs {
			close(out.Ch)
		}
		return nil
	}

	// OLS : slope = (n*sumXY - sumX*sumY) / (n*sumX2 - sumX^2)
	var sumX, sumY, sumXY, sumX2, sumY2 float64
	for i := range xs {
		sumX += xs[i]
		sumY += ys[i]
		sumXY += xs[i] * ys[i]
		sumX2 += xs[i] * xs[i]
		sumY2 += ys[i] * ys[i]
	}
	slope := (n*sumXY - sumX*sumY) / (n*sumX2 - sumX*sumX)
	intercept := (sumY - slope*sumX) / n

	// R²
	yMean := sumY / n
	var ssTot, ssRes float64
	for i := range xs {
		pred := slope*xs[i] + intercept
		ssRes += math.Pow(ys[i]-pred, 2)
		ssTot += math.Pow(ys[i]-yMean, 2)
	}
	r2 := 1.0
	if ssTot != 0 {
		r2 = 1 - ssRes/ssTot
	}

	for _, row := range rows {
		newRow := make(contracts.DataRow, len(row)+3)
		for k, v := range row {
			newRow[k] = v
		}
		if x, err := strconv.ParseFloat(fmt.Sprintf("%v", row[feature]), 64); err == nil {
			newRow[outCol] = slope*x + intercept
		}
		newRow["slope"] = fmt.Sprintf("%.6f", slope)
		newRow["intercept"] = fmt.Sprintf("%.6f", intercept)
		newRow["r2"] = fmt.Sprintf("%.4f", r2)
		for _, out := range bctx.Outputs {
			out.Ch <- newRow
		}
	}
	for _, out := range bctx.Outputs {
		close(out.Ch)
	}
	return nil
}
