package ml

import (
	"fmt"
	"math"
	"strconv"

	"github.com/rinjold/go-etl-studio/internal/etl/blocks"
	"github.com/rinjold/go-etl-studio/internal/etl/contracts"
)

func init() {
	blocks.Register("ml.naive_bayes", func() contracts.Block { return &NaiveBayes{} })
}

// NaiveBayes effectue une classification Naive Bayes gaussienne.
// Paramètres :
//
//	features : colonnes numériques séparées par virgule (ex: "age,salary")
//	target   : colonne de classe
//	output   : colonne de prédiction (défaut: "predicted_class")
type NaiveBayes struct{}

func (b *NaiveBayes) Type() string { return "ml.naive_bayes" }

func (b *NaiveBayes) Run(bctx *contracts.BlockContext) error {
	featuresParam := bctx.Params["features"]
	target := bctx.Params["target"]
	outCol := bctx.Params["output"]
	if featuresParam == "" || target == "" {
		return fmt.Errorf("ml.naive_bayes: 'features' et 'target' obligatoires")
	}
	if outCol == "" {
		outCol = "predicted_class"
	}

	var featureCols []string
	for _, f := range splitTrim(featuresParam) {
		featureCols = append(featureCols, f)
	}

	// Phase d'entraînement : collecter toutes les lignes.
	type classStat struct {
		count int
		sums  map[string]float64
		sumSq map[string]float64
	}
	classStats := make(map[string]*classStat)
	var rows []contracts.DataRow
	var totalRows int

	for row := range bctx.Inputs[0].Ch {
		rows = append(rows, row)
		totalRows++
		cls := fmt.Sprintf("%v", row[target])
		if _, ok := classStats[cls]; !ok {
			classStats[cls] = &classStat{
				sums:  make(map[string]float64),
				sumSq: make(map[string]float64),
			}
		}
		cs := classStats[cls]
		cs.count++
		for _, f := range featureCols {
			val, _ := strconv.ParseFloat(fmt.Sprintf("%v", row[f]), 64)
			cs.sums[f] += val
			cs.sumSq[f] += val * val
		}
	}

	// Calculer mean et variance par classe et feature.
	type gaussParam struct{ mean, variance float64 }
	params := make(map[string]map[string]gaussParam)
	for cls, cs := range classStats {
		params[cls] = make(map[string]gaussParam)
		n := float64(cs.count)
		for _, f := range featureCols {
			mean := cs.sums[f] / n
			variance := cs.sumSq[f]/n - mean*mean
			if variance < 1e-9 {
				variance = 1e-9
			}
			params[cls][f] = gaussParam{mean, variance}
		}
	}

	gaussLogProb := func(x, mean, variance float64) float64 {
		return -0.5*math.Log(2*math.Pi*variance) - math.Pow(x-mean, 2)/(2*variance)
	}

	// Prédiction sur chaque ligne.
	for _, row := range rows {
		bestClass := ""
		bestScore := math.Inf(-1)
		for cls, cs := range classStats {
			logProb := math.Log(float64(cs.count) / float64(totalRows))
			for _, f := range featureCols {
				val, _ := strconv.ParseFloat(fmt.Sprintf("%v", row[f]), 64)
				gp := params[cls][f]
				logProb += gaussLogProb(val, gp.mean, gp.variance)
			}
			if logProb > bestScore {
				bestScore = logProb
				bestClass = cls
			}
		}
		newRow := make(contracts.DataRow, len(row)+1)
		for k, v := range row {
			newRow[k] = v
		}
		newRow[outCol] = bestClass
		for _, out := range bctx.Outputs {
			out.Ch <- newRow
		}
	}
	for _, out := range bctx.Outputs {
		close(out.Ch)
	}
	return nil
}

func splitTrim(s string) []string {
	var result []string
	for _, p := range splitComma(s) {
		p = trimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}

func splitComma(s string) []string {
	var parts []string
	start := 0
	for i, c := range s {
		if c == ',' {
			parts = append(parts, s[start:i])
			start = i + 1
		}
	}
	parts = append(parts, s[start:])
	return parts
}

func trimSpace(s string) string {
	start, end := 0, len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t') {
		end--
	}
	return s[start:end]
}
