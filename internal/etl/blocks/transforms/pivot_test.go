package transforms

import (
	"context"
	"testing"

	"github.com/rinjold/go-etl-studio/internal/etl/contracts"
)

func TestPivot_SumAggregation(t *testing.T) {
	rows := []contracts.DataRow{
		{"region": "Nord", "product": "A", "amount": "100"},
		{"region": "Nord", "product": "B", "amount": "200"},
		{"region": "Sud", "product": "A", "amount": "150"},
		{"region": "Nord", "product": "A", "amount": "50"}, // doublon Nord/A → SUM=150
	}
	out := &contracts.Port{Ch: make(chan contracts.DataRow, 10)}
	bctx := &contracts.BlockContext{
		Ctx: context.Background(),
		Params: map[string]string{
			"groupBy":     "region",
			"pivotColumn": "product",
			"valueColumn": "amount",
			"aggregation": "SUM",
		},
		Inputs:  []*contracts.Port{makePort(rows)},
		Outputs: []*contracts.Port{out},
	}
	if err := (&Pivot{}).Run(bctx); err != nil {
		t.Fatalf("pivot SUM: %v", err)
	}
	result := collectOutput(out)
	if len(result) != 2 {
		t.Fatalf("attendu 2 lignes (Nord, Sud), obtenu %d", len(result))
	}
	// Groupes triés: Nord[0], Sud[1]
	nord := result[0]
	if nord["region"] != "Nord" {
		t.Errorf("result[0] doit être Nord, obtenu %v", nord["region"])
	}
	if nord["A"] != 150.0 {
		t.Errorf("Nord/A SUM: attendu 150, obtenu %v", nord["A"])
	}
	if nord["B"] != 200.0 {
		t.Errorf("Nord/B SUM: attendu 200, obtenu %v", nord["B"])
	}
	sud := result[1]
	if sud["B"] != 0.0 {
		t.Errorf("Sud/B: aucune valeur, attendu 0, obtenu %v", sud["B"])
	}
}

func TestPivot_CountAggregation(t *testing.T) {
	rows := []contracts.DataRow{
		{"dept": "IT", "cat": "bug", "id": "1"},
		{"dept": "IT", "cat": "bug", "id": "2"},
		{"dept": "IT", "cat": "feature", "id": "3"},
	}
	out := &contracts.Port{Ch: make(chan contracts.DataRow, 10)}
	bctx := &contracts.BlockContext{
		Ctx: context.Background(),
		Params: map[string]string{
			"groupBy":     "dept",
			"pivotColumn": "cat",
			"valueColumn": "id",
			"aggregation": "COUNT",
		},
		Inputs:  []*contracts.Port{makePort(rows)},
		Outputs: []*contracts.Port{out},
	}
	if err := (&Pivot{}).Run(bctx); err != nil {
		t.Fatalf("pivot COUNT: %v", err)
	}
	result := collectOutput(out)
	if result[0]["bug"] != 2.0 {
		t.Errorf("bug COUNT: attendu 2, obtenu %v", result[0]["bug"])
	}
	if result[0]["feature"] != 1.0 {
		t.Errorf("feature COUNT: attendu 1, obtenu %v", result[0]["feature"])
	}
}

func TestPivot_MissingParams(t *testing.T) {
	out := &contracts.Port{Ch: make(chan contracts.DataRow, 1)}
	bctx := &contracts.BlockContext{
		Ctx:     context.Background(),
		Params:  map[string]string{"groupBy": "x", "pivotColumn": "y"},
		Inputs:  []*contracts.Port{makePort(nil)},
		Outputs: []*contracts.Port{out},
	}
	if err := (&Pivot{}).Run(bctx); err == nil {
		t.Error("valueColumn manquant: attendu une erreur")
	}
}
