package transforms

import (
	"context"
	"testing"

	"github.com/rinjold/go-etl-studio/internal/etl/contracts"
)

func makePort(rows []contracts.DataRow) *contracts.Port {
	ch := make(chan contracts.DataRow, len(rows)+1)
	for _, r := range rows {
		ch <- r
	}
	close(ch)
	return &contracts.Port{Ch: ch}
}

func collectOutput(out *contracts.Port) []contracts.DataRow {
	var result []contracts.DataRow
	for row := range out.Ch {
		result = append(result, row)
	}
	return result
}

// --- Tests Merge ---

func TestMerge_InnerJoin(t *testing.T) {
	left := []contracts.DataRow{
		{"id": "1", "name": "Alice"},
		{"id": "2", "name": "Bob"},
		{"id": "3", "name": "Carol"},
	}
	right := []contracts.DataRow{
		{"id": "1", "dept": "Engineering"},
		{"id": "2", "dept": "Marketing"},
	}

	out := &contracts.Port{Ch: make(chan contracts.DataRow, 10)}
	bctx := &contracts.BlockContext{
		Ctx:     context.Background(),
		Params:  map[string]string{"on": "id", "how": "inner"},
		Inputs:  []*contracts.Port{makePort(left), makePort(right)},
		Outputs: []*contracts.Port{out},
	}

	b := &Merge{}
	if err := b.Run(bctx); err != nil {
		t.Fatalf("inner join error: %v", err)
	}
	rows := collectOutput(out)
	if len(rows) != 2 {
		t.Errorf("inner join: attendu 2 lignes, obtenu %d", len(rows))
	}
}

func TestMerge_LeftJoin(t *testing.T) {
	left := []contracts.DataRow{
		{"id": "1", "name": "Alice"},
		{"id": "99", "name": "Ghost"},
	}
	right := []contracts.DataRow{
		{"id": "1", "dept": "Engineering"},
	}

	out := &contracts.Port{Ch: make(chan contracts.DataRow, 10)}
	bctx := &contracts.BlockContext{
		Ctx:     context.Background(),
		Params:  map[string]string{"on": "id", "how": "left"},
		Inputs:  []*contracts.Port{makePort(left), makePort(right)},
		Outputs: []*contracts.Port{out},
	}

	b := &Merge{}
	if err := b.Run(bctx); err != nil {
		t.Fatalf("left join error: %v", err)
	}
	rows := collectOutput(out)
	if len(rows) != 2 {
		t.Errorf("left join: attendu 2 lignes, obtenu %d", len(rows))
	}
}

func TestMerge_Suffixes(t *testing.T) {
	left := []contracts.DataRow{
		{"id": "1", "value": "left_val"},
	}
	right := []contracts.DataRow{
		{"id": "1", "value": "right_val"},
	}

	out := &contracts.Port{Ch: make(chan contracts.DataRow, 10)}
	bctx := &contracts.BlockContext{
		Ctx: context.Background(),
		Params: map[string]string{
			"on":           "id",
			"how":          "inner",
			"left_suffix":  "_left",
			"right_suffix": "_right",
		},
		Inputs:  []*contracts.Port{makePort(left), makePort(right)},
		Outputs: []*contracts.Port{out},
	}

	b := &Merge{}
	if err := b.Run(bctx); err != nil {
		t.Fatalf("suffixes error: %v", err)
	}
	rows := collectOutput(out)
	if len(rows) != 1 {
		t.Fatalf("suffixes: attendu 1 ligne, obtenu %d", len(rows))
	}
	if rows[0]["value_left"] == nil {
		t.Errorf("suffixes: colonne 'value_left' manquante, got: %v", rows[0])
	}
	if rows[0]["value_right"] == nil {
		t.Errorf("suffixes: colonne 'value_right' manquante, got: %v", rows[0])
	}
}

func TestMerge_ValidateOneToOne_Fail(t *testing.T) {
	left := []contracts.DataRow{
		{"id": "1", "name": "Alice"},
	}
	// Clé droite dupliquée → doit échouer avec validate=one_to_one
	right := []contracts.DataRow{
		{"id": "1", "dept": "Eng"},
		{"id": "1", "dept": "Mkt"},
	}

	out := &contracts.Port{Ch: make(chan contracts.DataRow, 10)}
	bctx := &contracts.BlockContext{
		Ctx: context.Background(),
		Params: map[string]string{
			"on":       "id",
			"how":      "inner",
			"validate": "one_to_one",
		},
		Inputs:  []*contracts.Port{makePort(left), makePort(right)},
		Outputs: []*contracts.Port{out},
	}

	b := &Merge{}
	if err := b.Run(bctx); err == nil {
		t.Error("validate one_to_one: attendu une erreur, aucune reçue")
	}
}

func TestMerge_OuterJoin(t *testing.T) {
	left := []contracts.DataRow{
		{"id": "1", "name": "Alice"},
	}
	right := []contracts.DataRow{
		{"id": "2", "dept": "HR"},
	}

	out := &contracts.Port{Ch: make(chan contracts.DataRow, 10)}
	bctx := &contracts.BlockContext{
		Ctx:     context.Background(),
		Params:  map[string]string{"on": "id", "how": "outer"},
		Inputs:  []*contracts.Port{makePort(left), makePort(right)},
		Outputs: []*contracts.Port{out},
	}

	b := &Merge{}
	if err := b.Run(bctx); err != nil {
		t.Fatalf("outer join error: %v", err)
	}
	rows := collectOutput(out)
	if len(rows) != 2 {
		t.Errorf("outer join: attendu 2 lignes, obtenu %d", len(rows))
	}
}

func TestMerge_MissingKey(t *testing.T) {
	out := &contracts.Port{Ch: make(chan contracts.DataRow, 1)}
	bctx := &contracts.BlockContext{
		Ctx:     context.Background(),
		Params:  map[string]string{"how": "inner"},
		Inputs:  []*contracts.Port{makePort(nil), makePort(nil)},
		Outputs: []*contracts.Port{out},
	}
	b := &Merge{}
	if err := b.Run(bctx); err == nil {
		t.Error("clé manquante: attendu une erreur")
	}
}
