package transforms

import (
	"context"
	"testing"

	"github.com/rinjold/go-etl-studio/internal/etl/contracts"
)

func TestSplit_RouteToFirstMatch(t *testing.T) {
	rows := []contracts.DataRow{
		{"amount": "1500"},
		{"amount": "700"},
		{"amount": "100"},
	}
	out0 := &contracts.Port{Ch: make(chan contracts.DataRow, 5)}
	out1 := &contracts.Port{Ch: make(chan contracts.DataRow, 5)}
	out2 := &contracts.Port{Ch: make(chan contracts.DataRow, 5)}
	bctx := &contracts.BlockContext{
		Ctx:    context.Background(),
		Params: map[string]string{"conditions": "amount > 1000, amount > 500"},
		Inputs: []*contracts.Port{makePort(rows)},
		Outputs: []*contracts.Port{out0, out1, out2},
	}
	if err := (&Split{}).Run(bctx); err != nil {
		t.Fatalf("split route: %v", err)
	}
	r0 := collectOutput(out0)
	r1 := collectOutput(out1)
	r2 := collectOutput(out2)
	if len(r0) != 1 {
		t.Errorf("out0 (>1000): attendu 1, obtenu %d", len(r0))
	}
	if len(r1) != 1 {
		t.Errorf("out1 (>500): attendu 1, obtenu %d", len(r1))
	}
	if len(r2) != 1 {
		t.Errorf("out2 (catch-all): attendu 1, obtenu %d", len(r2))
	}
}

func TestSplit_AllCatchAll(t *testing.T) {
	rows := []contracts.DataRow{
		{"val": "1"},
		{"val": "2"},
	}
	out0 := &contracts.Port{Ch: make(chan contracts.DataRow, 5)}
	out1 := &contracts.Port{Ch: make(chan contracts.DataRow, 5)}
	bctx := &contracts.BlockContext{
		Ctx:     context.Background(),
		Params:  map[string]string{"conditions": "val > 1000"},
		Inputs:  []*contracts.Port{makePort(rows)},
		Outputs: []*contracts.Port{out0, out1},
	}
	if err := (&Split{}).Run(bctx); err != nil {
		t.Fatalf("split catch-all: %v", err)
	}
	r0 := collectOutput(out0)
	r1 := collectOutput(out1)
	if len(r0) != 0 {
		t.Errorf("out0: aucune ligne attendue, obtenu %d", len(r0))
	}
	if len(r1) != 2 {
		t.Errorf("catch-all: attendu 2 lignes, obtenu %d", len(r1))
	}
}

func TestSplit_ConditionsCountMismatch(t *testing.T) {
	out0 := &contracts.Port{Ch: make(chan contracts.DataRow, 1)}
	out1 := &contracts.Port{Ch: make(chan contracts.DataRow, 1)}
	out2 := &contracts.Port{Ch: make(chan contracts.DataRow, 1)}
	// 1 condition mais 3 sorties (2 attendues)
	bctx := &contracts.BlockContext{
		Ctx:     context.Background(),
		Params:  map[string]string{"conditions": "val > 10"},
		Inputs:  []*contracts.Port{makePort(nil)},
		Outputs: []*contracts.Port{out0, out1, out2},
	}
	if err := (&Split{}).Run(bctx); err == nil {
		t.Error("count mismatch: attendu une erreur")
	}
}

func TestSplit_MissingConditionsParam(t *testing.T) {
	out0 := &contracts.Port{Ch: make(chan contracts.DataRow, 1)}
	out1 := &contracts.Port{Ch: make(chan contracts.DataRow, 1)}
	bctx := &contracts.BlockContext{
		Ctx:     context.Background(),
		Params:  map[string]string{},
		Inputs:  []*contracts.Port{makePort(nil)},
		Outputs: []*contracts.Port{out0, out1},
	}
	if err := (&Split{}).Run(bctx); err == nil {
		t.Error("conditions manquantes: attendu une erreur")
	}
}
