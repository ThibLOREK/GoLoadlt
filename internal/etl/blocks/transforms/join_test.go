package transforms

import (
	"context"
	"testing"

	"github.com/rinjold/go-etl-studio/internal/etl/contracts"
)

func TestJoin_Inner(t *testing.T) {
	left := []contracts.DataRow{
		{"user_id": "1", "name": "Alice"},
		{"user_id": "2", "name": "Bob"},
		{"user_id": "3", "name": "Carol"},
	}
	right := []contracts.DataRow{
		{"id": "1", "score": "90"},
		{"id": "2", "score": "80"},
		// id=3 absent → Carol exclue en inner
	}
	out := &contracts.Port{Ch: make(chan contracts.DataRow, 10)}
	bctx := &contracts.BlockContext{
		Ctx: context.Background(),
		Params: map[string]string{
			"leftKey":  "user_id",
			"rightKey": "id",
			"type":     "inner",
		},
		Inputs:  []*contracts.Port{makePort(left), makePort(right)},
		Outputs: []*contracts.Port{out},
	}
	if err := (&Join{}).Run(bctx); err != nil {
		t.Fatalf("join inner: %v", err)
	}
	result := collectOutput(out)
	if len(result) != 2 {
		t.Errorf("inner join: attendu 2 lignes, obtenu %d", len(result))
	}
}

func TestJoin_Left(t *testing.T) {
	left := []contracts.DataRow{
		{"user_id": "1", "name": "Alice"},
		{"user_id": "99", "name": "Ghost"}, // pas de match
	}
	right := []contracts.DataRow{
		{"id": "1", "score": "90"},
	}
	out := &contracts.Port{Ch: make(chan contracts.DataRow, 10)}
	bctx := &contracts.BlockContext{
		Ctx: context.Background(),
		Params: map[string]string{
			"leftKey":  "user_id",
			"rightKey": "id",
			"type":     "left",
		},
		Inputs:  []*contracts.Port{makePort(left), makePort(right)},
		Outputs: []*contracts.Port{out},
	}
	if err := (&Join{}).Run(bctx); err != nil {
		t.Fatalf("join left: %v", err)
	}
	result := collectOutput(out)
	if len(result) != 2 {
		t.Errorf("left join: attendu 2 lignes (1 match + 1 unmatched), obtenu %d", len(result))
	}
}

func TestJoin_DuplicatedColumnPrefixed(t *testing.T) {
	// Si left et right ont une colonne du même nom (hors rightKey), le droit est préfixé "right_"
	left := []contracts.DataRow{
		{"id": "1", "name": "Alice"},
	}
	right := []contracts.DataRow{
		{"id": "1", "name": "AliceR", "score": "99"},
	}
	out := &contracts.Port{Ch: make(chan contracts.DataRow, 5)}
	bctx := &contracts.BlockContext{
		Ctx: context.Background(),
		Params: map[string]string{
			"leftKey":  "id",
			"rightKey": "id",
			"type":     "inner",
		},
		Inputs:  []*contracts.Port{makePort(left), makePort(right)},
		Outputs: []*contracts.Port{out},
	}
	if err := (&Join{}).Run(bctx); err != nil {
		t.Fatalf("join prefix: %v", err)
	}
	result := collectOutput(out)
	if result[0]["name"] != "Alice" {
		t.Errorf("colonne gauche 'name' doit être Alice, obtenu %v", result[0]["name"])
	}
	if result[0]["right_name"] != "AliceR" {
		t.Errorf("colonne droite doit être préfixée 'right_name', obtenu %v", result[0]["right_name"])
	}
	if result[0]["score"] != "99" {
		t.Errorf("score doit être présent, obtenu %v", result[0]["score"])
	}
}

func TestJoin_NotEnoughInputs(t *testing.T) {
	out := &contracts.Port{Ch: make(chan contracts.DataRow, 1)}
	bctx := &contracts.BlockContext{
		Ctx:     context.Background(),
		Params:  map[string]string{"leftKey": "id", "rightKey": "id"},
		Inputs:  []*contracts.Port{makePort(nil)},
		Outputs: []*contracts.Port{out},
	}
	if err := (&Join{}).Run(bctx); err == nil {
		t.Error("1 seul port: attendu une erreur")
	}
}

func TestJoin_InvalidType(t *testing.T) {
	out := &contracts.Port{Ch: make(chan contracts.DataRow, 1)}
	bctx := &contracts.BlockContext{
		Ctx: context.Background(),
		Params: map[string]string{
			"leftKey":  "id",
			"rightKey": "id",
			"type":     "cross",
		},
		Inputs:  []*contracts.Port{makePort(nil), makePort(nil)},
		Outputs: []*contracts.Port{out},
	}
	if err := (&Join{}).Run(bctx); err == nil {
		t.Error("type 'cross' non supporté: attendu une erreur")
	}
}
