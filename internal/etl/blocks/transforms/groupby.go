package transforms

import (
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"

	"github.com/rinjold/go-etl-studio/internal/etl/blocks"
	"github.com/rinjold/go-etl-studio/internal/etl/contracts"
)

func init() {
	blocks.Register("transform.groupby", func() contracts.Block { return &GroupBy{} })
}

// GroupBy agrège un flux par groupe avec parité complète df.groupby().
//
// Paramètres (bctx.Params) :
//   - by          : colonnes de groupe séparées par virgule (ex: "region,category")
//   - aggregations: liste CSV de "FUNC(col)" ou "FUNC(col) AS alias"
//                   Fonctions supportées : SUM, COUNT, AVG, MIN, MAX,
//                   FIRST, LAST, MEDIAN, NUNIQUE, STD, VAR
//                   Exemple : "SUM(amount),COUNT(id) AS nb,AVG(price) AS avg_price"
//   - sort        : "true" (défaut) | "false" — trier les groupes sur les clés
//   - as_index    : "true" (défaut) | "false" — inclure les colonnes 'by' dans la sortie
//   - dropna      : "true" (défaut) | "false" — exclure les lignes où 'by' contient ""/null
type GroupBy struct{}

func (b *GroupBy) Type() string { return "transform.groupby" }

type gbAggSpec struct {
	func_  string
	col    string
	alias  string
}

// parseGBAggSpecs parse "SUM(amount),COUNT(id) AS nb,AVG(price) AS avg_price"
func parseGBAggSpecs(raw string) ([]gbAggSpec, error) {
	var specs []gbAggSpec
	for _, part := range splitComma(raw) {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		// Extraire alias optionnel : "SUM(x) AS total"
		alias := ""
		if idx := strings.Index(strings.ToUpper(part), " AS "); idx >= 0 {
			alias = strings.TrimSpace(part[idx+4:])
			part = strings.TrimSpace(part[:idx])
		}
		open := strings.Index(part, "(")
		close_ := strings.LastIndex(part, ")")
		if open < 0 || close_ < 0 || close_ <= open {
			return nil, fmt.Errorf("transform.groupby: agrégation invalide '%s'", part)
		}
		fn := strings.ToUpper(strings.TrimSpace(part[:open]))
		col := strings.TrimSpace(part[open+1 : close_])
		if alias == "" {
			alias = fn + "_" + col
		}
		switch fn {
		case "SUM", "COUNT", "AVG", "MIN", "MAX", "FIRST", "LAST", "MEDIAN", "NUNIQUE", "STD", "VAR":
		default:
			return nil, fmt.Errorf("transform.groupby: fonction '%s' non supportée", fn)
		}
		specs = append(specs, gbAggSpec{func_: fn, col: col, alias: alias})
	}
	return specs, nil
}

func (b *GroupBy) Run(bctx *contracts.BlockContext) error {
	byCSV := bctx.Params["by"]
	aggsRaw := bctx.Params["aggregations"]
	if byCSV == "" || aggsRaw == "" {
		return fmt.Errorf("transform.groupby: paramètres 'by' et 'aggregations' requis")
	}
	if len(bctx.Inputs) == 0 {
		return fmt.Errorf("transform.groupby: aucun port d'entrée")
	}

	byCols := splitComma(byCSV)
	specs, err := parseGBAggSpecs(aggsRaw)
	if err != nil {
		return err
	}

	sortGroups := bctx.Params["sort"] != "false"
	asIndex := bctx.Params["as_index"] != "false"
	dropNA := bctx.Params["dropna"] != "false"

	// --- Accumulation ---
	type groupAcc struct {
		values map[string][]float64   // pour SUM/AVG/MIN/MAX/STD/VAR/MEDIAN
		strVals map[string][]string   // pour FIRST/LAST/NUNIQUE
		counts  map[string]int        // pour COUNT
	}

	groups := make(map[string]*groupAcc)
	var groupKeys []string
	seen := make(map[string]bool)

	for row := range bctx.Inputs[0].Ch {
		// Construction de la clé de groupe
		parts := make([]string, len(byCols))
		skip := false
		for i, c := range byCols {
			v := fmt.Sprintf("%v", row[c])
			if dropNA && (v == "" || v == "<nil>") {
				skip = true
				break
			}
			parts[i] = v
		}
		if skip {
			continue
		}
		gk := strings.Join(parts, "\x00")

		if !seen[gk] {
			seen[gk] = true
			groupKeys = append(groupKeys, gk)
			groups[gk] = &groupAcc{
				values:  make(map[string][]float64),
				strVals: make(map[string][]string),
				counts:  make(map[string]int),
			}
		}
		acc := groups[gk]

		for _, spec := range specs {
			rawVal := fmt.Sprintf("%v", row[spec.col])
			switch spec.func_ {
			case "FIRST":
				if len(acc.strVals[spec.alias]) == 0 {
					acc.strVals[spec.alias] = []string{rawVal}
				}
			case "LAST":
				acc.strVals[spec.alias] = []string{rawVal}
			case "NUNIQUE":
				acc.strVals[spec.alias] = append(acc.strVals[spec.alias], rawVal)
			case "COUNT":
				if rawVal != "" && rawVal != "<nil>" {
					acc.counts[spec.alias]++
				}
			default:
				v, _ := strconv.ParseFloat(rawVal, 64)
				acc.values[spec.alias] = append(acc.values[spec.alias], v)
			}
		}
	}

	// --- Tri optionnel des groupes ---
	if sortGroups {
		sort.Strings(groupKeys)
	}

	closeOutputs := func() {
		for _, out := range bctx.Outputs {
			close(out.Ch)
		}
	}

	// --- Émission des résultats ---
	for _, gk := range groupKeys {
		keyParts := strings.Split(gk, "\x00")
		row := make(contracts.DataRow)

		if asIndex {
			for i, c := range byCols {
				row[c] = keyParts[i]
			}
		}

		acc := groups[gk]
		for _, spec := range specs {
			switch spec.func_ {
			case "FIRST", "LAST":
				if len(acc.strVals[spec.alias]) > 0 {
					row[spec.alias] = acc.strVals[spec.alias][0]
				} else {
					row[spec.alias] = nil
				}
			case "NUNIQUE":
				uniq := make(map[string]struct{})
				for _, sv := range acc.strVals[spec.alias] {
					uniq[sv] = struct{}{}
				}
				row[spec.alias] = len(uniq)
			case "COUNT":
				row[spec.alias] = acc.counts[spec.alias]
			default:
				row[spec.alias] = computeAgg(spec.func_, acc.values[spec.alias])
			}
		}

		for _, out := range bctx.Outputs {
			select {
			case out.Ch <- row:
			case <-bctx.Ctx.Done():
				closeOutputs()
				return bctx.Ctx.Err()
			}
		}
	}
	closeOutputs()
	return nil
}

// computeAgg calcule l'agrégat numérique pour un groupe.
func computeAgg(fn string, vals []float64) any {
	if len(vals) == 0 {
		return nil
	}
	switch fn {
	case "SUM":
		var s float64
		for _, v := range vals {
			s += v
		}
		return s
	case "AVG":
		var s float64
		for _, v := range vals {
			s += v
		}
		return s / float64(len(vals))
	case "MIN":
		m := vals[0]
		for _, v := range vals[1:] {
			if v < m {
				m = v
			}
		}
		return m
	case "MAX":
		m := vals[0]
		for _, v := range vals[1:] {
			if v > m {
				m = v
			}
		}
		return m
	case "MEDIAN":
		sorted := make([]float64, len(vals))
		copy(sorted, vals)
		sort.Float64s(sorted)
		n := len(sorted)
		if n%2 == 0 {
			return (sorted[n/2-1] + sorted[n/2]) / 2
		}
		return sorted[n/2]
	case "STD":
		return math.Sqrt(computeVariance(vals))
	case "VAR":
		return computeVariance(vals)
	}
	return nil
}

func computeVariance(vals []float64) float64 {
	if len(vals) < 2 {
		return 0
	}
	var sum float64
	for _, v := range vals {
		sum += v
	}
	mean := sum / float64(len(vals))
	var variance float64
	for _, v := range vals {
		d := v - mean
		variance += d * d
	}
	return variance / float64(len(vals)-1) // écart-type de Bessel (sample)
}
