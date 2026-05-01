package transforms

import (
	"context"
	"testing"

	"github.com/rinjold/go-etl-studio/internal/etl/contracts"
)

func TestDedup_CSVKeys(t *testing.T) {
	rows := []contracts.DataRow{
		{"id": "1", "country": "FR", "val": "a"},
		{"id": "1", "country": "FR", "val": "b"}, // doublon sur id+country
		{"id": "2", "country": "FR", "val": "c"},
	}
	out := &contracts.Port{Ch: make(chan contracts.DataRow, 10)}
	bctx := &contracts.BlockContext{
		Ctx:     context.Background(),
		Params:  map[string]string{"keys": "id,country"},
		Inputs:  []*contracts.Port{makePort(rows)},
		Outputs: []*contracts.Port{out},
	}
	if err := (&Dedup{}).Run(bctx); err != nil {
		t.Fatalf("dedup CSV: %v", err)
	}
	result := collectOutput(out)
	if len(result) != 2 {
		t.Errorf("dedup CSV: attendu 2 lignes, obtenu %d", len(result))
	}
}

func TestDedup_IndexedKeys(t *testing.T) {
	rows := []contracts.DataRow{
		{"a": "x", "b": "1"},
		{"a": "x", "b": "2"}, // doublon sur key_0=a
		{"a": "y", "b": "1"},
	}
	out := &contracts.Port{Ch: make(chan contracts.DataRow, 10)}
	bctx := &contracts.BlockContext{
		Ctx:     context.Background(),
		Params:  map[string]string{"key_0": "a"},
		Inputs:  []*contracts.Port{makePort(rows)},
		Outputs: []*contracts.Port{out},
	}
	if err := (&Dedup{}).Run(bctx); err != nil {
		t.Fatalf("dedup indexed: %v", err)
	}
	result := collectOutput(out)
	if len(result) != 2 {
		t.Errorf("dedup indexed: attendu 2 lignes, obtenu %d", len(result))
	}
}

func TestDedup_JSONArrayKeys(t *testing.T) {
	rows := []contracts.DataRow{
		{"x": "1", "y": "A", "z": "p"},
		{"x": "1", "y": "A", "z": "q"}, // doublon x+y
		{"x": "1", "y": "B", "z": "r"},
	}
	out := &contracts.Port{Ch: make(chan contracts.DataRow, 10)}
	bctx := &contracts.BlockContext{
		Ctx:     context.Background(),
		Params:  map[string]string{"keys": `["x","y"]`},
		Inputs:  []*contracts.Port{makePort(rows)},
		Outputs: []*contracts.Port{out},
	}
	if err := (&Dedup{}).Run(bctx); err != nil {
		t.Fatalf("dedup JSON: %v", err)
	}
	result := collectOutput(out)
	if len(result) != 2 {
		t.Errorf("dedup JSON: attendu 2 lignes, obtenu %d", len(result))
	}
}

func TestDedup_NoKeys_AllColumns(t *testing.T) {
	rows := []contracts.DataRow{
		{"a": "1", "b": "2"},
		{"a": "1", "b": "2"}, // doublon total
		{"a": "1", "b": "3"},
	}
	out := &contracts.Port{Ch: make(chan contracts.DataRow, 10)}
	bctx := &contracts.BlockContext{
		Ctx:     context.Background(),
		Params:  map[string]string{},
		Inputs:  []*contracts.Port{makePort(rows)},
		Outputs: []*contracts.Port{out},
	}
	if err := (&Dedup{}).Run(bctx); err != nil {
		t.Fatalf("dedup all cols: %v", err)
	}
	result := collectOutput(out)
	if len(result) != 2 {
		t.Errorf("dedup all cols: attendu 2 lignes, obtenu %d", len(result))
	}
}

func TestDedup_NoDuplicates(t *testing.T) {
	rows := []contracts.DataRow{
		{"id": "1"},
		{"id": "2"},
		{"id": "3"},
	}
	out := &contracts.Port{Ch: make(chan contracts.DataRow, 10)}
	bctx := &contracts.BlockContext{
		Ctx:     context.Background(),
		Params:  map[string]string{"keys": "id"},
		Inputs:  []*contracts.Port{makePort(rows)},
		Outputs: []*contracts.Port{out},
	}
	if err := (&Dedup{}).Run(bctx); err != nil {
		t.Fatalf("dedup no dup: %v", err)
	}
	result := collectOutput(out)
	if len(result) != 3 {
		t.Errorf("dedup no dup: attendu 3 lignes, obtenu %d", len(result))
	}
}
