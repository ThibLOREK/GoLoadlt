package transforms

import (
	"context"
	"testing"

	"github.com/rinjold/go-etl-studio/internal/etl/contracts"
)

func TestAggregate_SumCount(t *testing.T) {
	rows := []contracts.DataRow{
		{"region": "Nord", "amount": "100", "id": "1"},
		{"region": "Nord", "amount": "200", "id": "2"},
		{"region": "Sud", "amount": "150", "id": "3"},
	}
	out := &contracts.Port{Ch: make(chan contracts.DataRow, 10)}
	bctx := &contracts.BlockContext{
		Ctx: context.Background(),
		Params: map[string]string{
			"groupBy":      "region",
			"aggregations": "SUM(amount),COUNT(id)",
		},
		Inputs:  []*contracts.Port{makePort(rows)},
		Outputs: []*contracts.Port{out},
	}
	if err := (&Aggregate{}).Run(bctx); err != nil {
		t.Fatalf("aggregate sum/count: %v", err)
	}
	result := collectOutput(out)
	if len(result) != 2 {
		t.Fatalf("attendu 2 groupes, obtenu %d", len(result))
	}
	// Groupes triés alphabétiquement: Nord avant Sud
	nord := result[0]
	if nord["SUM_amount"] != 300.0 {
		t.Errorf("Nord SUM_amount: attendu 300, obtenu %v", nord["SUM_amount"])
	}
	if nord["COUNT_id"] != 2.0 {
		t.Errorf("Nord COUNT_id: attendu 2, obtenu %v", nord["COUNT_id"])
	}
}

func TestAggregate_AvgMinMax(t *testing.T) {
	rows := []contracts.DataRow{
		{"cat": "A", "val": "10"},
		{"cat": "A", "val": "30"},
		{"cat": "A", "val": "20"},
	}
	out := &contracts.Port{Ch: make(chan contracts.DataRow, 10)}
	bctx := &contracts.BlockContext{
		Ctx: context.Background(),
		Params: map[string]string{
			"groupBy":      "cat",
			"aggregations": "AVG(val),MIN(val),MAX(val)",
		},
		Inputs:  []*contracts.Port{makePort(rows)},
		Outputs: []*contracts.Port{out},
	}
	if err := (&Aggregate{}).Run(bctx); err != nil {
		t.Fatalf("aggregate avg/min/max: %v", err)
	}
	result := collectOutput(out)
	if result[0]["AVG_val"] != 20.0 {
		t.Errorf("AVG_val: attendu 20, obtenu %v", result[0]["AVG_val"])
	}
	if result[0]["MIN_val"] != 10.0 {
		t.Errorf("MIN_val: attendu 10, obtenu %v", result[0]["MIN_val"])
	}
	if result[0]["MAX_val"] != 30.0 {
		t.Errorf("MAX_val: attendu 30, obtenu %v", result[0]["MAX_val"])
	}
}

func TestAggregate_MultiGroupBy(t *testing.T) {
	rows := []contracts.DataRow{
		{"region": "Nord", "cat": "A", "val": "10"},
		{"region": "Nord", "cat": "A", "val": "20"},
		{"region": "Nord", "cat": "B", "val": "5"},
		{"region": "Sud", "cat": "A", "val": "30"},
	}
	out := &contracts.Port{Ch: make(chan contracts.DataRow, 10)}
	bctx := &contracts.BlockContext{
		Ctx: context.Background(),
		Params: map[string]string{
			"groupBy":      "region,cat",
			"aggregations": "SUM(val)",
		},
		Inputs:  []*contracts.Port{makePort(rows)},
		Outputs: []*contracts.Port{out},
	}
	if err := (&Aggregate{}).Run(bctx); err != nil {
		t.Fatalf("aggregate multi-groupby: %v", err)
	}
	result := collectOutput(out)
	if len(result) != 3 {
		t.Errorf("multi-groupby: attendu 3 groupes, obtenu %d", len(result))
	}
}

func TestAggregate_MissingParams(t *testing.T) {
	out := &contracts.Port{Ch: make(chan contracts.DataRow, 1)}
	bctx := &contracts.BlockContext{
		Ctx:     context.Background(),
		Params:  map[string]string{"groupBy": "x"},
		Inputs:  []*contracts.Port{makePort(nil)},
		Outputs: []*contracts.Port{out},
	}
	if err := (&Aggregate{}).Run(bctx); err == nil {
		t.Error("aggregations manquant: attendu une erreur")
	}
}
