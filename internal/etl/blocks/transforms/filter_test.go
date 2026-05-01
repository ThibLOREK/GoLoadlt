package transforms

import (
	"context"
	"testing"

	"github.com/rinjold/go-etl-studio/internal/etl/contracts"
)

func TestFilter_MatchAll(t *testing.T) {
	rows := []contracts.DataRow{
		{"amount": "150", "country": "FR"},
		{"amount": "50", "country": "DE"},
		{"amount": "200", "country": "FR"},
	}
	out := &contracts.Port{Ch: make(chan contracts.DataRow, 10)}
	bctx := &contracts.BlockContext{
		Ctx:     context.Background(),
		Params:  map[string]string{"condition": "amount > 100"},
		Inputs:  []*contracts.Port{makePort(rows)},
		Outputs: []*contracts.Port{out},
	}
	if err := (&Filter{}).Run(bctx); err != nil {
		t.Fatalf("filter: %v", err)
	}
	result := collectOutput(out)
	if len(result) != 2 {
		t.Errorf("attendu 2 lignes filtrées, obtenu %d", len(result))
	}
}

func TestFilter_NoMatch(t *testing.T) {
	rows := []contracts.DataRow{
		{"amount": "10"},
		{"amount": "20"},
	}
	out := &contracts.Port{Ch: make(chan contracts.DataRow, 5)}
	bctx := &contracts.BlockContext{
		Ctx:     context.Background(),
		Params:  map[string]string{"condition": "amount > 100"},
		Inputs:  []*contracts.Port{makePort(rows)},
		Outputs: []*contracts.Port{out},
	}
	if err := (&Filter{}).Run(bctx); err != nil {
		t.Fatalf("filter no-match: %v", err)
	}
	result := collectOutput(out)
	if len(result) != 0 {
		t.Errorf("attendu 0 lignes, obtenu %d", len(result))
	}
}

func TestFilter_StringEquality(t *testing.T) {
	rows := []contracts.DataRow{
		{"country": "FR", "val": "1"},
		{"country": "DE", "val": "2"},
		{"country": "FR", "val": "3"},
	}
	out := &contracts.Port{Ch: make(chan contracts.DataRow, 10)}
	bctx := &contracts.BlockContext{
		Ctx:     context.Background(),
		Params:  map[string]string{"condition": "country == 'FR'"},
		Inputs:  []*contracts.Port{makePort(rows)},
		Outputs: []*contracts.Port{out},
	}
	if err := (&Filter{}).Run(bctx); err != nil {
		t.Fatalf("filter string eq: %v", err)
	}
	result := collectOutput(out)
	if len(result) != 2 {
		t.Errorf("attendu 2 lignes FR, obtenu %d", len(result))
	}
}

func TestFilter_MissingConditionParam(t *testing.T) {
	out := &contracts.Port{Ch: make(chan contracts.DataRow, 1)}
	bctx := &contracts.BlockContext{
		Ctx:     context.Background(),
		Params:  map[string]string{},
		Inputs:  []*contracts.Port{makePort(nil)},
		Outputs: []*contracts.Port{out},
	}
	if err := (&Filter{}).Run(bctx); err == nil {
		t.Error("condition manquante: attendu une erreur")
	}
}

func TestFilter_NoInput(t *testing.T) {
	out := &contracts.Port{Ch: make(chan contracts.DataRow, 1)}
	bctx := &contracts.BlockContext{
		Ctx:     context.Background(),
		Params:  map[string]string{"condition": "amount > 0"},
		Inputs:  []*contracts.Port{},
		Outputs: []*contracts.Port{out},
	}
	if err := (&Filter{}).Run(bctx); err == nil {
		t.Error("aucun port d'entrée: attendu une erreur")
	}
}
