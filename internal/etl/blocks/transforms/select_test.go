package transforms

import (
	"context"
	"testing"

	"github.com/rinjold/go-etl-studio/internal/etl/contracts"
)

func TestSelect_SubsetColumns(t *testing.T) {
	rows := []contracts.DataRow{
		{"id": "1", "name": "Alice", "age": "30", "city": "Paris"},
		{"id": "2", "name": "Bob", "age": "25", "city": "Lyon"},
	}
	out := &contracts.Port{Ch: make(chan contracts.DataRow, 10)}
	bctx := &contracts.BlockContext{
		Ctx:     context.Background(),
		Params:  map[string]string{"columns": "id,name"},
		Inputs:  []*contracts.Port{makePort(rows)},
		Outputs: []*contracts.Port{out},
	}
	if err := (&Select{}).Run(bctx); err != nil {
		t.Fatalf("select subset: %v", err)
	}
	result := collectOutput(out)
	if len(result) != 2 {
		t.Fatalf("attendu 2 lignes, obtenu %d", len(result))
	}
	if _, exists := result[0]["age"]; exists {
		t.Error("colonne 'age' doit être supprimée")
	}
	if _, exists := result[0]["city"]; exists {
		t.Error("colonne 'city' doit être supprimée")
	}
	if result[0]["id"] != "1" {
		t.Errorf("id: attendu '1', obtenu %v", result[0]["id"])
	}
}

func TestSelect_UnknownColumnReturnsNil(t *testing.T) {
	rows := []contracts.DataRow{
		{"id": "1"},
	}
	out := &contracts.Port{Ch: make(chan contracts.DataRow, 5)}
	bctx := &contracts.BlockContext{
		Ctx:     context.Background(),
		Params:  map[string]string{"columns": "id,inexistant"},
		Inputs:  []*contracts.Port{makePort(rows)},
		Outputs: []*contracts.Port{out},
	}
	if err := (&Select{}).Run(bctx); err != nil {
		t.Fatalf("select colonne inconnue: %v", err)
	}
	result := collectOutput(out)
	// colonne inconnue retourne nil (zero value map)
	if result[0]["inexistant"] != nil {
		t.Errorf("colonne inconnue: attendu nil, obtenu %v", result[0]["inexistant"])
	}
}

func TestSelect_SingleColumn(t *testing.T) {
	rows := []contracts.DataRow{
		{"a": "1", "b": "2", "c": "3"},
	}
	out := &contracts.Port{Ch: make(chan contracts.DataRow, 5)}
	bctx := &contracts.BlockContext{
		Ctx:     context.Background(),
		Params:  map[string]string{"columns": "b"},
		Inputs:  []*contracts.Port{makePort(rows)},
		Outputs: []*contracts.Port{out},
	}
	if err := (&Select{}).Run(bctx); err != nil {
		t.Fatalf("select single col: %v", err)
	}
	result := collectOutput(out)
	if len(result[0]) != 1 {
		t.Errorf("attendu 1 colonne, obtenu %d", len(result[0]))
	}
	if result[0]["b"] != "2" {
		t.Errorf("select single: attendu '2', obtenu %v", result[0]["b"])
	}
}

func TestSelect_MissingColumnsParam(t *testing.T) {
	out := &contracts.Port{Ch: make(chan contracts.DataRow, 1)}
	bctx := &contracts.BlockContext{
		Ctx:     context.Background(),
		Params:  map[string]string{},
		Inputs:  []*contracts.Port{makePort(nil)},
		Outputs: []*contracts.Port{out},
	}
	if err := (&Select{}).Run(bctx); err == nil {
		t.Error("columns manquant: attendu une erreur")
	}
}
