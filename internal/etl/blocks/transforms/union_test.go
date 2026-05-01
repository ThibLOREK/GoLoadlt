package transforms

import (
	"context"
	"testing"

	"github.com/rinjold/go-etl-studio/internal/etl/contracts"
)

func TestUnion_Basic(t *testing.T) {
	rows1 := []contracts.DataRow{
		{"id": "1", "val": "a"},
		{"id": "2", "val": "b"},
	}
	rows2 := []contracts.DataRow{
		{"id": "3", "val": "c"},
		{"id": "4", "val": "d"},
	}
	out := &contracts.Port{Ch: make(chan contracts.DataRow, 10)}
	bctx := &contracts.BlockContext{
		Ctx:     context.Background(),
		Params:  map[string]string{},
		Inputs:  []*contracts.Port{makePort(rows1), makePort(rows2)},
		Outputs: []*contracts.Port{out},
	}
	if err := (&Union{}).Run(bctx); err != nil {
		t.Fatalf("union basic: %v", err)
	}
	result := collectOutput(out)
	if len(result) != 4 {
		t.Errorf("union: attendu 4 lignes, obtenu %d", len(result))
	}
}

func TestUnion_DifferentColumns(t *testing.T) {
	// Union de deux flux avec colonnes différentes : les absentes doivent être nil
	rows1 := []contracts.DataRow{
		{"id": "1", "name": "Alice"},
	}
	rows2 := []contracts.DataRow{
		{"id": "2", "score": "99"},
	}
	out := &contracts.Port{Ch: make(chan contracts.DataRow, 10)}
	bctx := &contracts.BlockContext{
		Ctx:     context.Background(),
		Params:  map[string]string{},
		Inputs:  []*contracts.Port{makePort(rows1), makePort(rows2)},
		Outputs: []*contracts.Port{out},
	}
	if err := (&Union{}).Run(bctx); err != nil {
		t.Fatalf("union diff cols: %v", err)
	}
	result := collectOutput(out)
	if len(result) != 2 {
		t.Errorf("union diff cols: attendu 2 lignes, obtenu %d", len(result))
	}
}

func TestUnion_SingleInput(t *testing.T) {
	rows := []contracts.DataRow{
		{"x": "1"},
		{"x": "2"},
	}
	out := &contracts.Port{Ch: make(chan contracts.DataRow, 10)}
	bctx := &contracts.BlockContext{
		Ctx:     context.Background(),
		Params:  map[string]string{},
		Inputs:  []*contracts.Port{makePort(rows)},
		Outputs: []*contracts.Port{out},
	}
	if err := (&Union{}).Run(bctx); err != nil {
		t.Fatalf("union single input: %v", err)
	}
	result := collectOutput(out)
	if len(result) != 2 {
		t.Errorf("union single: attendu 2 lignes, obtenu %d", len(result))
	}
}

func TestUnion_NoInput(t *testing.T) {
	out := &contracts.Port{Ch: make(chan contracts.DataRow, 1)}
	bctx := &contracts.BlockContext{
		Ctx:     context.Background(),
		Params:  map[string]string{},
		Inputs:  []*contracts.Port{},
		Outputs: []*contracts.Port{out},
	}
	if err := (&Union{}).Run(bctx); err == nil {
		t.Error("aucun port: attendu une erreur")
	}
}
