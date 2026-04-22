package transforms

import (
	"context"
	"testing"

	"github.com/rinjold/go-etl-studio/internal/etl/contracts"
)

func TestFillna_Scalar(t *testing.T) {
	rows := []contracts.DataRow{
		{"name": "Alice", "score": ""},
		{"name": "", "score": "95"},
		{"name": "Carol", "score": "80"},
	}
	out := &contracts.Port{Ch: make(chan contracts.DataRow, 10)}
	bctx := &contracts.BlockContext{
		Ctx:     context.Background(),
		Params:  map[string]string{"value": "0"},
		Inputs:  []*contracts.Port{makePort(rows)},
		Outputs: []*contracts.Port{out},
	}
	b := &Fillna{}
	if err := b.Run(bctx); err != nil {
		t.Fatalf("Fillna scalar: %v", err)
	}
	result := collectOutput(out)
	if result[0]["score"] != "0" {
		t.Errorf("score vide: attendu '0', obtenu %v", result[0]["score"])
	}
	if result[1]["name"] != "0" {
		t.Errorf("name vide: attendu '0', obtenu %v", result[1]["name"])
	}
	if result[2]["name"] != "Carol" {
		t.Errorf("name non-vide: ne doit pas être modifié, obtenu %v", result[2]["name"])
	}
}

func TestFillna_ScalarColumns(t *testing.T) {
	rows := []contracts.DataRow{
		{"a": "", "b": "", "c": "ok"},
	}
	out := &contracts.Port{Ch: make(chan contracts.DataRow, 10)}
	bctx := &contracts.BlockContext{
		Ctx:    context.Background(),
		Params: map[string]string{"value": "X", "columns": "a"},
		Inputs:  []*contracts.Port{makePort(rows)},
		Outputs: []*contracts.Port{out},
	}
	b := &Fillna{}
	if err := b.Run(bctx); err != nil {
		t.Fatalf("Fillna columns: %v", err)
	}
	result := collectOutput(out)
	if result[0]["a"] != "X" {
		t.Errorf("colonne 'a': attendu 'X', obtenu %v", result[0]["a"])
	}
	if result[0]["b"] != "" {
		t.Errorf("colonne 'b' ne devrait pas être modifiée, obtenu %v", result[0]["b"])
	}
}

func TestFillna_Ffill(t *testing.T) {
	rows := []contracts.DataRow{
		{"v": "A"},
		{"v": ""},
		{"v": ""},
		{"v": "B"},
		{"v": ""},
	}
	out := &contracts.Port{Ch: make(chan contracts.DataRow, 10)}
	bctx := &contracts.BlockContext{
		Ctx:    context.Background(),
		Params: map[string]string{"method": "ffill"},
		Inputs:  []*contracts.Port{makePort(rows)},
		Outputs: []*contracts.Port{out},
	}
	b := &Fillna{}
	if err := b.Run(bctx); err != nil {
		t.Fatalf("Fillna ffill: %v", err)
	}
	result := collectOutput(out)
	expected := []string{"A", "A", "A", "B", "B"}
	for i, exp := range expected {
		if result[i]["v"] != exp {
			t.Errorf("ffill[%d]: attendu '%s', obtenu %v", i, exp, result[i]["v"])
		}
	}
}

func TestFillna_FfillLimit(t *testing.T) {
	rows := []contracts.DataRow{
		{"v": "A"},
		{"v": ""},
		{"v": ""},
		{"v": ""},
	}
	out := &contracts.Port{Ch: make(chan contracts.DataRow, 10)}
	bctx := &contracts.BlockContext{
		Ctx:    context.Background(),
		Params: map[string]string{"method": "ffill", "limit": "1"},
		Inputs:  []*contracts.Port{makePort(rows)},
		Outputs: []*contracts.Port{out},
	}
	b := &Fillna{}
	if err := b.Run(bctx); err != nil {
		t.Fatalf("Fillna ffill limit: %v", err)
	}
	result := collectOutput(out)
	// Seule la 1ère valeur nulle doit être remplie
	if result[1]["v"] != "A" {
		t.Errorf("ffill limit[1]: attendu 'A', obtenu %v", result[1]["v"])
	}
	if result[2]["v"] != "" {
		t.Errorf("ffill limit[2]: doit rester vide, obtenu %v", result[2]["v"])
	}
}

func TestFillna_Bfill(t *testing.T) {
	rows := []contracts.DataRow{
		{"v": ""},
		{"v": ""},
		{"v": "Z"},
	}
	out := &contracts.Port{Ch: make(chan contracts.DataRow, 10)}
	bctx := &contracts.BlockContext{
		Ctx:    context.Background(),
		Params: map[string]string{"method": "bfill"},
		Inputs:  []*contracts.Port{makePort(rows)},
		Outputs: []*contracts.Port{out},
	}
	b := &Fillna{}
	if err := b.Run(bctx); err != nil {
		t.Fatalf("Fillna bfill: %v", err)
	}
	result := collectOutput(out)
	if result[0]["v"] != "Z" {
		t.Errorf("bfill[0]: attendu 'Z', obtenu %v", result[0]["v"])
	}
	if result[1]["v"] != "Z" {
		t.Errorf("bfill[1]: attendu 'Z', obtenu %v", result[1]["v"])
	}
}

func TestFillna_InvalidMethod(t *testing.T) {
	out := &contracts.Port{Ch: make(chan contracts.DataRow, 1)}
	bctx := &contracts.BlockContext{
		Ctx:     context.Background(),
		Params:  map[string]string{"method": "pad"},
		Inputs:  []*contracts.Port{makePort(nil)},
		Outputs: []*contracts.Port{out},
	}
	b := &Fillna{}
	if err := b.Run(bctx); err == nil {
		t.Error("method invalide: attendu une erreur")
	}
}
