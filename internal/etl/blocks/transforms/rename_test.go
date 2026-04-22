package transforms

import (
	"context"
	"testing"

	"github.com/rinjold/go-etl-studio/internal/etl/contracts"
)

func TestRename_Basic(t *testing.T) {
	rows := []contracts.DataRow{
		{"user_id": "1", "first_name": "Alice", "score": "90"},
	}
	out := &contracts.Port{Ch: make(chan contracts.DataRow, 5)}
	bctx := &contracts.BlockContext{
		Ctx:     context.Background(),
		Params:  map[string]string{"columns": "user_id:id,first_name:prenom"},
		Inputs:  []*contracts.Port{makePort(rows)},
		Outputs: []*contracts.Port{out},
	}
	b := &Rename{}
	if err := b.Run(bctx); err != nil {
		t.Fatalf("Rename basic: %v", err)
	}
	result := collectOutput(out)
	if result[0]["id"] != "1" {
		t.Errorf("user_id→id: attendu '1', obtenu %v", result[0]["id"])
	}
	if result[0]["prenom"] != "Alice" {
		t.Errorf("first_name→prenom: attendu 'Alice', obtenu %v", result[0]["prenom"])
	}
	if _, exists := result[0]["user_id"]; exists {
		t.Error("user_id ne doit plus exister après renommage")
	}
	if result[0]["score"] != "90" {
		t.Errorf("score non renommé: doit être préservé, obtenu %v", result[0]["score"])
	}
}

func TestRename_ErrorsRaise(t *testing.T) {
	rows := []contracts.DataRow{
		{"a": "1"},
	}
	out := &contracts.Port{Ch: make(chan contracts.DataRow, 5)}
	bctx := &contracts.BlockContext{
		Ctx:     context.Background(),
		Params:  map[string]string{"columns": "inexistant:nouveau", "errors": "raise"},
		Inputs:  []*contracts.Port{makePort(rows)},
		Outputs: []*contracts.Port{out},
	}
	b := &Rename{}
	if err := b.Run(bctx); err == nil {
		t.Error("errors=raise avec colonne manquante: attendu une erreur")
	}
}

func TestRename_ErrorsIgnore(t *testing.T) {
	rows := []contracts.DataRow{
		{"a": "1"},
	}
	out := &contracts.Port{Ch: make(chan contracts.DataRow, 5)}
	bctx := &contracts.BlockContext{
		Ctx:     context.Background(),
		Params:  map[string]string{"columns": "inexistant:nouveau", "errors": "ignore"},
		Inputs:  []*contracts.Port{makePort(rows)},
		Outputs: []*contracts.Port{out},
	}
	b := &Rename{}
	if err := b.Run(bctx); err != nil {
		t.Fatalf("errors=ignore: ne doit pas retourner d'erreur, got %v", err)
	}
	result := collectOutput(out)
	if result[0]["a"] != "1" {
		t.Errorf("colonne 'a' préservée: attendu '1', obtenu %v", result[0]["a"])
	}
}

func TestRename_MissingColumnsParam(t *testing.T) {
	out := &contracts.Port{Ch: make(chan contracts.DataRow, 1)}
	bctx := &contracts.BlockContext{
		Ctx:     context.Background(),
		Params:  map[string]string{},
		Inputs:  []*contracts.Port{makePort(nil)},
		Outputs: []*contracts.Port{out},
	}
	b := &Rename{}
	if err := b.Run(bctx); err == nil {
		t.Error("columns manquant: attendu une erreur")
	}
}
