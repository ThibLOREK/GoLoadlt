package transforms

import (
	"context"
	"testing"

	"github.com/rinjold/go-etl-studio/internal/etl/contracts"
)

func TestAddColumn_ArithmeticExpression(t *testing.T) {
	rows := []contracts.DataRow{
		{"amount": "100", "tax": "20"},
		{"amount": "200", "tax": "40"},
	}
	out := &contracts.Port{Ch: make(chan contracts.DataRow, 10)}
	bctx := &contracts.BlockContext{
		Ctx:     context.Background(),
		Params:  map[string]string{"name": "total", "expression": "amount + tax"},
		Inputs:  []*contracts.Port{makePort(rows)},
		Outputs: []*contracts.Port{out},
	}
	if err := (&AddColumn{}).Run(bctx); err != nil {
		t.Fatalf("add_column arithmetic: %v", err)
	}
	result := collectOutput(out)
	if len(result) != 2 {
		t.Fatalf("attendu 2 lignes, obtenu %d", len(result))
	}
	// La colonne source doit être préservée.
	if result[0]["amount"] != "100" {
		t.Errorf("amount doit être préservé, obtenu %v", result[0]["amount"])
	}
	// La nouvelle colonne doit exister.
	if _, exists := result[0]["total"]; !exists {
		t.Error("colonne 'total' doit exister")
	}
}

func TestAddColumn_ColumnCopy(t *testing.T) {
	rows := []contracts.DataRow{
		{"src": "hello"},
	}
	out := &contracts.Port{Ch: make(chan contracts.DataRow, 5)}
	bctx := &contracts.BlockContext{
		Ctx:     context.Background(),
		Params:  map[string]string{"name": "copy", "expression": "src"},
		Inputs:  []*contracts.Port{makePort(rows)},
		Outputs: []*contracts.Port{out},
	}
	if err := (&AddColumn{}).Run(bctx); err != nil {
		t.Fatalf("add_column copy: %v", err)
	}
	result := collectOutput(out)
	if result[0]["copy"] != result[0]["src"] {
		t.Errorf("copie colonne: copy=%v != src=%v", result[0]["copy"], result[0]["src"])
	}
}

func TestAddColumn_MissingNameParam(t *testing.T) {
	out := &contracts.Port{Ch: make(chan contracts.DataRow, 1)}
	bctx := &contracts.BlockContext{
		Ctx:     context.Background(),
		Params:  map[string]string{"expression": "amount * 2"},
		Inputs:  []*contracts.Port{makePort(nil)},
		Outputs: []*contracts.Port{out},
	}
	if err := (&AddColumn{}).Run(bctx); err == nil {
		t.Error("name manquant: attendu une erreur")
	}
}

func TestAddColumn_MissingExpressionParam(t *testing.T) {
	out := &contracts.Port{Ch: make(chan contracts.DataRow, 1)}
	bctx := &contracts.BlockContext{
		Ctx:     context.Background(),
		Params:  map[string]string{"name": "total"},
		Inputs:  []*contracts.Port{makePort(nil)},
		Outputs: []*contracts.Port{out},
	}
	if err := (&AddColumn{}).Run(bctx); err == nil {
		t.Error("expression manquante: attendu une erreur")
	}
}
