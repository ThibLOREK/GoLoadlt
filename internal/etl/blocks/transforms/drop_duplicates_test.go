package transforms

import (
	"context"
	"testing"

	"github.com/rinjold/go-etl-studio/internal/etl/contracts"
)

func TestDropDuplicates_KeepFirst(t *testing.T) {
	rows := []contracts.DataRow{
		{"id": "1", "val": "A"},
		{"id": "1", "val": "B"}, // doublon sur id
		{"id": "2", "val": "C"},
	}
	out := &contracts.Port{Ch: make(chan contracts.DataRow, 10)}
	bctx := &contracts.BlockContext{
		Ctx:     context.Background(),
		Params:  map[string]string{"subset": "id", "keep": "first"},
		Inputs:  []*contracts.Port{makePort(rows)},
		Outputs: []*contracts.Port{out},
	}
	b := &DropDuplicates{}
	if err := b.Run(bctx); err != nil {
		t.Fatalf("KeepFirst: %v", err)
	}
	result := collectOutput(out)
	if len(result) != 2 {
		t.Errorf("keep=first: attendu 2, obtenu %d", len(result))
	}
	if result[0]["val"] != "A" {
		t.Errorf("keep=first: attendu val='A' (premier), obtenu %v", result[0]["val"])
	}
}

func TestDropDuplicates_KeepLast(t *testing.T) {
	rows := []contracts.DataRow{
		{"id": "1", "val": "A"},
		{"id": "1", "val": "B"},
		{"id": "2", "val": "C"},
	}
	out := &contracts.Port{Ch: make(chan contracts.DataRow, 10)}
	bctx := &contracts.BlockContext{
		Ctx:     context.Background(),
		Params:  map[string]string{"subset": "id", "keep": "last"},
		Inputs:  []*contracts.Port{makePort(rows)},
		Outputs: []*contracts.Port{out},
	}
	b := &DropDuplicates{}
	if err := b.Run(bctx); err != nil {
		t.Fatalf("KeepLast: %v", err)
	}
	result := collectOutput(out)
	if len(result) != 2 {
		t.Errorf("keep=last: attendu 2, obtenu %d", len(result))
	}
	if result[0]["val"] != "B" {
		t.Errorf("keep=last: attendu val='B' (dernier), obtenu %v", result[0]["val"])
	}
}

func TestDropDuplicates_KeepFalse(t *testing.T) {
	rows := []contracts.DataRow{
		{"id": "1", "val": "A"},
		{"id": "1", "val": "B"}, // doublon → supprimé
		{"id": "2", "val": "C"}, // unique → conservé
	}
	out := &contracts.Port{Ch: make(chan contracts.DataRow, 10)}
	bctx := &contracts.BlockContext{
		Ctx:     context.Background(),
		Params:  map[string]string{"subset": "id", "keep": "false"},
		Inputs:  []*contracts.Port{makePort(rows)},
		Outputs: []*contracts.Port{out},
	}
	b := &DropDuplicates{}
	if err := b.Run(bctx); err != nil {
		t.Fatalf("KeepFalse: %v", err)
	}
	result := collectOutput(out)
	if len(result) != 1 {
		t.Errorf("keep=false: attendu 1 (unique seulement), obtenu %d", len(result))
	}
	if result[0]["id"] != "2" {
		t.Errorf("keep=false: attendu id='2', obtenu %v", result[0]["id"])
	}
}

func TestDropDuplicates_NoSubset(t *testing.T) {
	rows := []contracts.DataRow{
		{"a": "1", "b": "X"},
		{"a": "1", "b": "X"}, // doublon complet
		{"a": "1", "b": "Y"}, // différent sur b
	}
	out := &contracts.Port{Ch: make(chan contracts.DataRow, 10)}
	bctx := &contracts.BlockContext{
		Ctx:     context.Background(),
		Params:  map[string]string{},
		Inputs:  []*contracts.Port{makePort(rows)},
		Outputs: []*contracts.Port{out},
	}
	b := &DropDuplicates{}
	if err := b.Run(bctx); err != nil {
		t.Fatalf("NoSubset: %v", err)
	}
	result := collectOutput(out)
	if len(result) != 2 {
		t.Errorf("no subset: attendu 2 lignes distinctes, obtenu %d", len(result))
	}
}

func TestDropDuplicates_IgnoreIndex(t *testing.T) {
	rows := []contracts.DataRow{
		{"id": "A"},
		{"id": "A"},
		{"id": "B"},
	}
	out := &contracts.Port{Ch: make(chan contracts.DataRow, 10)}
	bctx := &contracts.BlockContext{
		Ctx:     context.Background(),
		Params:  map[string]string{"subset": "id", "ignore_index": "true"},
		Inputs:  []*contracts.Port{makePort(rows)},
		Outputs: []*contracts.Port{out},
	}
	b := &DropDuplicates{}
	if err := b.Run(bctx); err != nil {
		t.Fatalf("IgnoreIndex: %v", err)
	}
	result := collectOutput(out)
	if len(result) != 2 {
		t.Fatalf("ignore_index: attendu 2, obtenu %d", len(result))
	}
	if result[0]["_index"] != 0 {
		t.Errorf("ignore_index[0]: attendu 0, obtenu %v", result[0]["_index"])
	}
	if result[1]["_index"] != 1 {
		t.Errorf("ignore_index[1]: attendu 1, obtenu %v", result[1]["_index"])
	}
}

func TestDropDuplicates_InvalidKeep(t *testing.T) {
	out := &contracts.Port{Ch: make(chan contracts.DataRow, 1)}
	bctx := &contracts.BlockContext{
		Ctx:     context.Background(),
		Params:  map[string]string{"keep": "none"},
		Inputs:  []*contracts.Port{makePort(nil)},
		Outputs: []*contracts.Port{out},
	}
	b := &DropDuplicates{}
	if err := b.Run(bctx); err == nil {
		t.Error("keep invalide: attendu une erreur")
	}
}
