package sources

import (
	"context"
	"testing"

	"github.com/rinjold/go-etl-studio/internal/etl/contracts"
)

func makeOutput(buf int) *contracts.Port {
	return &contracts.Port{Ch: make(chan contracts.DataRow, buf)}
}

func collect(p *contracts.Port) []contracts.DataRow {
	var rows []contracts.DataRow
	for r := range p.Ch {
		rows = append(rows, r)
	}
	return rows
}

func TestDataGrid_Basic(t *testing.T) {
	out := makeOutput(10)
	bctx := &contracts.BlockContext{
		Ctx: context.Background(),
		Params: map[string]string{
			"columns": "id,name,amount",
			"rows":    `[["1","Alice","100"],["2","Bob","200"]]`,
		},
		Outputs: []*contracts.Port{out},
	}
	b := &DataGrid{}
	if err := b.Run(bctx); err != nil {
		t.Fatalf("DataGrid basic: %v", err)
	}
	result := collect(out)
	if len(result) != 2 {
		t.Fatalf("attendu 2 lignes, obtenu %d", len(result))
	}
	if result[0]["name"] != "Alice" {
		t.Errorf("ligne 0 name: attendu 'Alice', obtenu %v", result[0]["name"])
	}
	if result[1]["amount"] != "200" {
		t.Errorf("ligne 1 amount: attendu '200', obtenu %v", result[1]["amount"])
	}
}

func TestDataGrid_NumericValues(t *testing.T) {
	out := makeOutput(10)
	bctx := &contracts.BlockContext{
		Ctx: context.Background(),
		Params: map[string]string{
			"columns": "id,score",
			"rows":    `[[1, 95.5], [2, 87.0]]`,
		},
		Outputs: []*contracts.Port{out},
	}
	b := &DataGrid{}
	if err := b.Run(bctx); err != nil {
		t.Fatalf("DataGrid numeric: %v", err)
	}
	result := collect(out)
	if len(result) != 2 {
		t.Fatalf("attendu 2 lignes, obtenu %d", len(result))
	}
	// Les valeurs numériques JSON sont stockées telles quelles (float64)
	if result[0]["score"] != 95.5 {
		t.Errorf("ligne 0 score: attendu 95.5, obtenu %v", result[0]["score"])
	}
}

func TestDataGrid_SingleRow(t *testing.T) {
	out := makeOutput(5)
	bctx := &contracts.BlockContext{
		Ctx: context.Background(),
		Params: map[string]string{
			"columns": "x",
			"rows":    `[["hello"]]`,
		},
		Outputs: []*contracts.Port{out},
	}
	b := &DataGrid{}
	if err := b.Run(bctx); err != nil {
		t.Fatalf("DataGrid single row: %v", err)
	}
	result := collect(out)
	if len(result) != 1 || result[0]["x"] != "hello" {
		t.Errorf("single row: attendu x='hello', obtenu %v", result)
	}
}

func TestDataGrid_EmptyRows(t *testing.T) {
	out := makeOutput(5)
	bctx := &contracts.BlockContext{
		Ctx: context.Background(),
		Params: map[string]string{
			"columns": "a,b",
			"rows":    `[]`,
		},
		Outputs: []*contracts.Port{out},
	}
	b := &DataGrid{}
	if err := b.Run(bctx); err != nil {
		t.Fatalf("DataGrid empty rows: %v", err)
	}
	result := collect(out)
	if len(result) != 0 {
		t.Errorf("empty rows: attendu 0, obtenu %d", len(result))
	}
}

func TestDataGrid_ColumnMismatch(t *testing.T) {
	out := makeOutput(5)
	bctx := &contracts.BlockContext{
		Ctx: context.Background(),
		Params: map[string]string{
			"columns": "a,b,c",
			"rows":    `[["1","2"]]`, // 2 valeurs pour 3 colonnes
		},
		Outputs: []*contracts.Port{out},
	}
	b := &DataGrid{}
	if err := b.Run(bctx); err == nil {
		t.Error("mismatch colonnes: attendu une erreur")
	}
}

func TestDataGrid_MissingColumns(t *testing.T) {
	out := makeOutput(5)
	bctx := &contracts.BlockContext{
		Ctx: context.Background(),
		Params: map[string]string{
			"rows": `[["1"]]`,
		},
		Outputs: []*contracts.Port{out},
	}
	b := &DataGrid{}
	if err := b.Run(bctx); err == nil {
		t.Error("columns manquant: attendu une erreur")
	}
}

func TestDataGrid_InvalidJSON(t *testing.T) {
	out := makeOutput(5)
	bctx := &contracts.BlockContext{
		Ctx: context.Background(),
		Params: map[string]string{
			"columns": "a",
			"rows":    `not json`,
		},
		Outputs: []*contracts.Port{out},
	}
	b := &DataGrid{}
	if err := b.Run(bctx); err == nil {
		t.Error("JSON invalide: attendu une erreur")
	}
}

func TestDataGrid_MultipleOutputs(t *testing.T) {
	out1 := makeOutput(5)
	out2 := makeOutput(5)
	bctx := &contracts.BlockContext{
		Ctx: context.Background(),
		Params: map[string]string{
			"columns": "v",
			"rows":    `[["A"],["B"]]`,
		},
		Outputs: []*contracts.Port{out1, out2},
	}
	b := &DataGrid{}
	if err := b.Run(bctx); err != nil {
		t.Fatalf("multi outputs: %v", err)
	}
	r1 := collect(out1)
	r2 := collect(out2)
	if len(r1) != 2 || len(r2) != 2 {
		t.Errorf("multi outputs: attendu 2+2 lignes, obtenu %d+%d", len(r1), len(r2))
	}
}
