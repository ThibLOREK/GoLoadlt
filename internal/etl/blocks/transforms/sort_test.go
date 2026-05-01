package transforms

import (
	"context"
	"testing"

	"github.com/rinjold/go-etl-studio/internal/etl/contracts"
)

func TestSort_NumericAsc(t *testing.T) {
	rows := []contracts.DataRow{
		{"amount": "300", "id": "3"},
		{"amount": "100", "id": "1"},
		{"amount": "200", "id": "2"},
	}
	out := &contracts.Port{Ch: make(chan contracts.DataRow, 10)}
	bctx := &contracts.BlockContext{
		Ctx:     context.Background(),
		Params:  map[string]string{"columns": "amount", "order": "asc"},
		Inputs:  []*contracts.Port{makePort(rows)},
		Outputs: []*contracts.Port{out},
	}
	if err := (&Sort{}).Run(bctx); err != nil {
		t.Fatalf("sort asc: %v", err)
	}
	result := collectOutput(out)
	if result[0]["amount"] != "100" {
		t.Errorf("sort asc[0]: attendu '100', obtenu %v", result[0]["amount"])
	}
	if result[2]["amount"] != "300" {
		t.Errorf("sort asc[2]: attendu '300', obtenu %v", result[2]["amount"])
	}
}

func TestSort_NumericDesc(t *testing.T) {
	rows := []contracts.DataRow{
		{"val": "5"},
		{"val": "1"},
		{"val": "9"},
		{"val": "3"},
	}
	out := &contracts.Port{Ch: make(chan contracts.DataRow, 10)}
	bctx := &contracts.BlockContext{
		Ctx:     context.Background(),
		Params:  map[string]string{"columns": "val", "order": "desc"},
		Inputs:  []*contracts.Port{makePort(rows)},
		Outputs: []*contracts.Port{out},
	}
	if err := (&Sort{}).Run(bctx); err != nil {
		t.Fatalf("sort desc: %v", err)
	}
	result := collectOutput(out)
	if result[0]["val"] != "9" {
		t.Errorf("sort desc[0]: attendu '9', obtenu %v", result[0]["val"])
	}
	if result[3]["val"] != "1" {
		t.Errorf("sort desc[3]: attendu '1', obtenu %v", result[3]["val"])
	}
}

func TestSort_StringAlpha(t *testing.T) {
	rows := []contracts.DataRow{
		{"name": "Charlie"},
		{"name": "Alice"},
		{"name": "Bob"},
	}
	out := &contracts.Port{Ch: make(chan contracts.DataRow, 10)}
	bctx := &contracts.BlockContext{
		Ctx:     context.Background(),
		Params:  map[string]string{"columns": "name"},
		Inputs:  []*contracts.Port{makePort(rows)},
		Outputs: []*contracts.Port{out},
	}
	if err := (&Sort{}).Run(bctx); err != nil {
		t.Fatalf("sort alpha: %v", err)
	}
	result := collectOutput(out)
	if result[0]["name"] != "Alice" {
		t.Errorf("sort alpha[0]: attendu 'Alice', obtenu %v", result[0]["name"])
	}
}

func TestSort_MultiColumns(t *testing.T) {
	rows := []contracts.DataRow{
		{"dept": "B", "salary": "30"},
		{"dept": "A", "salary": "50"},
		{"dept": "A", "salary": "20"},
		{"dept": "B", "salary": "10"},
	}
	out := &contracts.Port{Ch: make(chan contracts.DataRow, 10)}
	bctx := &contracts.BlockContext{
		Ctx:     context.Background(),
		Params:  map[string]string{"columns": "dept,salary", "order": "asc"},
		Inputs:  []*contracts.Port{makePort(rows)},
		Outputs: []*contracts.Port{out},
	}
	if err := (&Sort{}).Run(bctx); err != nil {
		t.Fatalf("sort multi: %v", err)
	}
	result := collectOutput(out)
	if result[0]["dept"] != "A" || result[0]["salary"] != "20" {
		t.Errorf("sort multi[0]: attendu A/20, obtenu %v/%v", result[0]["dept"], result[0]["salary"])
	}
}

func TestSort_MissingColumnsParam(t *testing.T) {
	out := &contracts.Port{Ch: make(chan contracts.DataRow, 1)}
	bctx := &contracts.BlockContext{
		Ctx:     context.Background(),
		Params:  map[string]string{},
		Inputs:  []*contracts.Port{makePort(nil)},
		Outputs: []*contracts.Port{out},
	}
	if err := (&Sort{}).Run(bctx); err == nil {
		t.Error("columns manquant: attendu une erreur")
	}
}
