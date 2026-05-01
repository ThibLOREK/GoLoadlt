package transforms

import (
	"context"
	"testing"

	"github.com/rinjold/go-etl-studio/internal/etl/contracts"
)

func TestCast_ToInt(t *testing.T) {
	rows := []contracts.DataRow{
		{"score": "42", "name": "Alice"},
	}
	out := &contracts.Port{Ch: make(chan contracts.DataRow, 5)}
	bctx := &contracts.BlockContext{
		Ctx:     context.Background(),
		Params:  map[string]string{"column": "score", "targetType": "int"},
		Inputs:  []*contracts.Port{makePort(rows)},
		Outputs: []*contracts.Port{out},
	}
	if err := (&Cast{}).Run(bctx); err != nil {
		t.Fatalf("cast int: %v", err)
	}
	result := collectOutput(out)
	if result[0]["score"] != 42 {
		t.Errorf("cast int: attendu 42 (int), obtenu %v (%T)", result[0]["score"], result[0]["score"])
	}
	if result[0]["name"] != "Alice" {
		t.Errorf("colonne non castée doit être préservée")
	}
}

func TestCast_ToFloat(t *testing.T) {
	rows := []contracts.DataRow{
		{"price": "3.14"},
	}
	out := &contracts.Port{Ch: make(chan contracts.DataRow, 5)}
	bctx := &contracts.BlockContext{
		Ctx:     context.Background(),
		Params:  map[string]string{"column": "price", "targetType": "float"},
		Inputs:  []*contracts.Port{makePort(rows)},
		Outputs: []*contracts.Port{out},
	}
	if err := (&Cast{}).Run(bctx); err != nil {
		t.Fatalf("cast float: %v", err)
	}
	result := collectOutput(out)
	if result[0]["price"] != 3.14 {
		t.Errorf("cast float: attendu 3.14, obtenu %v", result[0]["price"])
	}
}

func TestCast_ToBool(t *testing.T) {
	rows := []contracts.DataRow{
		{"active": "true"},
		{"active": "false"},
	}
	out := &contracts.Port{Ch: make(chan contracts.DataRow, 5)}
	bctx := &contracts.BlockContext{
		Ctx:     context.Background(),
		Params:  map[string]string{"column": "active", "targetType": "bool"},
		Inputs:  []*contracts.Port{makePort(rows)},
		Outputs: []*contracts.Port{out},
	}
	if err := (&Cast{}).Run(bctx); err != nil {
		t.Fatalf("cast bool: %v", err)
	}
	result := collectOutput(out)
	if result[0]["active"] != true {
		t.Errorf("cast bool true: obtenu %v", result[0]["active"])
	}
	if result[1]["active"] != false {
		t.Errorf("cast bool false: obtenu %v", result[1]["active"])
	}
}

func TestCast_ToString(t *testing.T) {
	rows := []contracts.DataRow{
		{"count": 99},
	}
	out := &contracts.Port{Ch: make(chan contracts.DataRow, 5)}
	bctx := &contracts.BlockContext{
		Ctx:     context.Background(),
		Params:  map[string]string{"column": "count", "targetType": "string"},
		Inputs:  []*contracts.Port{makePort(rows)},
		Outputs: []*contracts.Port{out},
	}
	if err := (&Cast{}).Run(bctx); err != nil {
		t.Fatalf("cast string: %v", err)
	}
	result := collectOutput(out)
	if result[0]["count"] != "99" {
		t.Errorf("cast string: attendu '99', obtenu %v", result[0]["count"])
	}
}

func TestCast_UnknownType(t *testing.T) {
	rows := []contracts.DataRow{{"x": "1"}}
	out := &contracts.Port{Ch: make(chan contracts.DataRow, 5)}
	bctx := &contracts.BlockContext{
		Ctx:     context.Background(),
		Params:  map[string]string{"column": "x", "targetType": "datetime"},
		Inputs:  []*contracts.Port{makePort(rows)},
		Outputs: []*contracts.Port{out},
	}
	if err := (&Cast{}).Run(bctx); err == nil {
		t.Error("type inconnu: attendu une erreur")
	}
}

func TestCast_MissingParams(t *testing.T) {
	out := &contracts.Port{Ch: make(chan contracts.DataRow, 1)}
	bctx := &contracts.BlockContext{
		Ctx:     context.Background(),
		Params:  map[string]string{"column": "x"},
		Inputs:  []*contracts.Port{makePort(nil)},
		Outputs: []*contracts.Port{out},
	}
	if err := (&Cast{}).Run(bctx); err == nil {
		t.Error("targetType manquant: attendu une erreur")
	}
}
