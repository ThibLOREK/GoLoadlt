package transforms

import (
	"context"
	"testing"

	"github.com/rinjold/go-etl-studio/internal/etl/contracts"
)

func TestUnpivot_ThreeColumns(t *testing.T) {
	rows := []contracts.DataRow{
		{"region": "Nord", "jan": "100", "fev": "200", "mar": "150"},
	}
	out := &contracts.Port{Ch: make(chan contracts.DataRow, 10)}
	bctx := &contracts.BlockContext{
		Ctx: context.Background(),
		Params: map[string]string{
			"columns":   "jan,fev,mar",
			"keyName":   "mois",
			"valueName": "montant",
		},
		Inputs:  []*contracts.Port{makePort(rows)},
		Outputs: []*contracts.Port{out},
	}
	if err := (&Unpivot{}).Run(bctx); err != nil {
		t.Fatalf("unpivot: %v", err)
	}
	result := collectOutput(out)
	if len(result) != 3 {
		t.Fatalf("1 ligne source × 3 colonnes = 3 lignes attendues, obtenu %d", len(result))
	}
	// Vérifier que chaque ligne a la bonne structure
	for _, r := range result {
		if _, ok := r["mois"]; !ok {
			t.Error("colonne 'mois' manquante dans la sortie")
		}
		if _, ok := r["montant"]; !ok {
			t.Error("colonne 'montant' manquante dans la sortie")
		}
		if _, ok := r["jan"]; ok {
			t.Error("colonne source 'jan' ne doit pas apparaître dans la sortie")
		}
		if r["region"] != "Nord" {
			t.Errorf("colonne 'region' doit être préservée, obtenu %v", r["region"])
		}
	}
}

func TestUnpivot_MultipleSourceRows(t *testing.T) {
	rows := []contracts.DataRow{
		{"id": "1", "q1": "10", "q2": "20"},
		{"id": "2", "q1": "30", "q2": "40"},
	}
	out := &contracts.Port{Ch: make(chan contracts.DataRow, 20)}
	bctx := &contracts.BlockContext{
		Ctx: context.Background(),
		Params: map[string]string{
			"columns":   "q1,q2",
			"keyName":   "quarter",
			"valueName": "value",
		},
		Inputs:  []*contracts.Port{makePort(rows)},
		Outputs: []*contracts.Port{out},
	}
	if err := (&Unpivot{}).Run(bctx); err != nil {
		t.Fatalf("unpivot multi rows: %v", err)
	}
	result := collectOutput(out)
	if len(result) != 4 {
		t.Errorf("2 rows × 2 cols = 4 lignes, obtenu %d", len(result))
	}
}

func TestUnpivot_MissingParams(t *testing.T) {
	out := &contracts.Port{Ch: make(chan contracts.DataRow, 1)}
	bctx := &contracts.BlockContext{
		Ctx:     context.Background(),
		Params:  map[string]string{"columns": "a,b", "keyName": "k"},
		Inputs:  []*contracts.Port{makePort(nil)},
		Outputs: []*contracts.Port{out},
	}
	if err := (&Unpivot{}).Run(bctx); err == nil {
		t.Error("valueName manquant: attendu une erreur")
	}
}
