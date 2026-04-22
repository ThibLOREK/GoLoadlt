package transforms

import (
	"context"
	"testing"

	"github.com/rinjold/go-etl-studio/internal/etl/contracts"
)

func TestGroupBy_SumCount(t *testing.T) {
	rows := []contracts.DataRow{
		{"region": "Nord", "amount": "100", "id": "1"},
		{"region": "Nord", "amount": "200", "id": "2"},
		{"region": "Sud", "amount": "150", "id": "3"},
	}
	out := &contracts.Port{Ch: make(chan contracts.DataRow, 10)}
	bctx := &contracts.BlockContext{
		Ctx: context.Background(),
		Params: map[string]string{
			"by":           "region",
			"aggregations": "SUM(amount) AS total,COUNT(id) AS nb",
		},
		Inputs:  []*contracts.Port{makePort(rows)},
		Outputs: []*contracts.Port{out},
	}
	b := &GroupBy{}
	if err := b.Run(bctx); err != nil {
		t.Fatalf("GroupBy SumCount: %v", err)
	}
	result := collectOutput(out)
	if len(result) != 2 {
		t.Fatalf("attendu 2 groupes, obtenu %d", len(result))
	}
	// Groupes triés par défaut: Nord, Sud
	nord := result[0]
	if nord["total"] != 300.0 {
		t.Errorf("Nord total: attendu 300, obtenu %v", nord["total"])
	}
	if nord["nb"] != 2 {
		t.Errorf("Nord nb: attendu 2, obtenu %v", nord["nb"])
	}
}

func TestGroupBy_AvgMinMax(t *testing.T) {
	rows := []contracts.DataRow{
		{"cat": "A", "val": "10"},
		{"cat": "A", "val": "30"},
		{"cat": "B", "val": "5"},
	}
	out := &contracts.Port{Ch: make(chan contracts.DataRow, 10)}
	bctx := &contracts.BlockContext{
		Ctx: context.Background(),
		Params: map[string]string{
			"by":           "cat",
			"aggregations": "AVG(val) AS moyenne,MIN(val) AS mini,MAX(val) AS maxi",
		},
		Inputs:  []*contracts.Port{makePort(rows)},
		Outputs: []*contracts.Port{out},
	}
	b := &GroupBy{}
	if err := b.Run(bctx); err != nil {
		t.Fatalf("GroupBy AvgMinMax: %v", err)
	}
	result := collectOutput(out)
	if len(result) != 2 {
		t.Fatalf("attendu 2 groupes, obtenu %d", len(result))
	}
	a := result[0] // cat A (trié)
	if a["moyenne"] != 20.0 {
		t.Errorf("A moyenne: attendu 20, obtenu %v", a["moyenne"])
	}
	if a["mini"] != 10.0 {
		t.Errorf("A mini: attendu 10, obtenu %v", a["mini"])
	}
	if a["maxi"] != 30.0 {
		t.Errorf("A maxi: attendu 30, obtenu %v", a["maxi"])
	}
}

func TestGroupBy_Median(t *testing.T) {
	rows := []contracts.DataRow{
		{"g": "X", "v": "1"},
		{"g": "X", "v": "3"},
		{"g": "X", "v": "5"},
	}
	out := &contracts.Port{Ch: make(chan contracts.DataRow, 10)}
	bctx := &contracts.BlockContext{
		Ctx:     context.Background(),
		Params:  map[string]string{"by": "g", "aggregations": "MEDIAN(v) AS med"},
		Inputs:  []*contracts.Port{makePort(rows)},
		Outputs: []*contracts.Port{out},
	}
	b := &GroupBy{}
	if err := b.Run(bctx); err != nil {
		t.Fatalf("GroupBy Median: %v", err)
	}
	result := collectOutput(out)
	if result[0]["med"] != 3.0 {
		t.Errorf("median: attendu 3, obtenu %v", result[0]["med"])
	}
}

func TestGroupBy_NUnique(t *testing.T) {
	rows := []contracts.DataRow{
		{"dept": "IT", "city": "Paris"},
		{"dept": "IT", "city": "Lyon"},
		{"dept": "IT", "city": "Paris"},
	}
	out := &contracts.Port{Ch: make(chan contracts.DataRow, 10)}
	bctx := &contracts.BlockContext{
		Ctx:     context.Background(),
		Params:  map[string]string{"by": "dept", "aggregations": "NUNIQUE(city) AS uniq_cities"},
		Inputs:  []*contracts.Port{makePort(rows)},
		Outputs: []*contracts.Port{out},
	}
	b := &GroupBy{}
	if err := b.Run(bctx); err != nil {
		t.Fatalf("GroupBy NUnique: %v", err)
	}
	result := collectOutput(out)
	if result[0]["uniq_cities"] != 2 {
		t.Errorf("nunique: attendu 2, obtenu %v", result[0]["uniq_cities"])
	}
}

func TestGroupBy_DropNA(t *testing.T) {
	rows := []contracts.DataRow{
		{"region": "Nord", "val": "10"},
		{"region": "", "val": "99"},   // ligne NA — doit être ignorée
		{"region": "Sud", "val": "20"},
	}
	out := &contracts.Port{Ch: make(chan contracts.DataRow, 10)}
	bctx := &contracts.BlockContext{
		Ctx: context.Background(),
		Params: map[string]string{
			"by":           "region",
			"aggregations": "SUM(val) AS total",
			"dropna":       "true",
		},
		Inputs:  []*contracts.Port{makePort(rows)},
		Outputs: []*contracts.Port{out},
	}
	b := &GroupBy{}
	if err := b.Run(bctx); err != nil {
		t.Fatalf("GroupBy DropNA: %v", err)
	}
	result := collectOutput(out)
	if len(result) != 2 {
		t.Errorf("dropna=true: attendu 2 groupes, obtenu %d", len(result))
	}
}

func TestGroupBy_AsIndexFalse(t *testing.T) {
	rows := []contracts.DataRow{
		{"grp": "A", "val": "5"},
		{"grp": "B", "val": "10"},
	}
	out := &contracts.Port{Ch: make(chan contracts.DataRow, 10)}
	bctx := &contracts.BlockContext{
		Ctx: context.Background(),
		Params: map[string]string{
			"by":           "grp",
			"aggregations": "SUM(val) AS total",
			"as_index":     "false",
		},
		Inputs:  []*contracts.Port{makePort(rows)},
		Outputs: []*contracts.Port{out},
	}
	b := &GroupBy{}
	if err := b.Run(bctx); err != nil {
		t.Fatalf("GroupBy AsIndexFalse: %v", err)
	}
	result := collectOutput(out)
	// as_index=false : la colonne 'grp' NE doit PAS apparaître
	if _, exists := result[0]["grp"]; exists {
		t.Errorf("as_index=false: colonne 'grp' ne devrait pas être présente")
	}
}

func TestGroupBy_SortFalse(t *testing.T) {
	rows := []contracts.DataRow{
		{"cat": "Z", "v": "1"},
		{"cat": "A", "v": "2"},
	}
	out := &contracts.Port{Ch: make(chan contracts.DataRow, 10)}
	bctx := &contracts.BlockContext{
		Ctx: context.Background(),
		Params: map[string]string{
			"by":           "cat",
			"aggregations": "COUNT(v) AS nb",
			"sort":         "false",
		},
		Inputs:  []*contracts.Port{makePort(rows)},
		Outputs: []*contracts.Port{out},
	}
	b := &GroupBy{}
	if err := b.Run(bctx); err != nil {
		t.Fatalf("GroupBy SortFalse: %v", err)
	}
	result := collectOutput(out)
	// sort=false : ordre d'apparition préservé — Z apparaît en premier
	if result[0]["cat"] != "Z" {
		t.Errorf("sort=false: attendu 'Z' en premier, obtenu %v", result[0]["cat"])
	}
}

func TestGroupBy_MultiKey(t *testing.T) {
	rows := []contracts.DataRow{
		{"region": "Nord", "cat": "A", "val": "10"},
		{"region": "Nord", "cat": "A", "val": "20"},
		{"region": "Nord", "cat": "B", "val": "5"},
		{"region": "Sud", "cat": "A", "val": "30"},
	}
	out := &contracts.Port{Ch: make(chan contracts.DataRow, 10)}
	bctx := &contracts.BlockContext{
		Ctx:     context.Background(),
		Params:  map[string]string{"by": "region,cat", "aggregations": "SUM(val) AS total"},
		Inputs:  []*contracts.Port{makePort(rows)},
		Outputs: []*contracts.Port{out},
	}
	b := &GroupBy{}
	if err := b.Run(bctx); err != nil {
		t.Fatalf("GroupBy MultiKey: %v", err)
	}
	result := collectOutput(out)
	if len(result) != 3 {
		t.Errorf("multi-key: attendu 3 groupes, obtenu %d", len(result))
	}
}

func TestGroupBy_InvalidFunc(t *testing.T) {
	out := &contracts.Port{Ch: make(chan contracts.DataRow, 1)}
	bctx := &contracts.BlockContext{
		Ctx:     context.Background(),
		Params:  map[string]string{"by": "x", "aggregations": "INVALID(y)"},
		Inputs:  []*contracts.Port{makePort(nil)},
		Outputs: []*contracts.Port{out},
	}
	b := &GroupBy{}
	if err := b.Run(bctx); err == nil {
		t.Error("fonction invalide: attendu une erreur")
	}
}
