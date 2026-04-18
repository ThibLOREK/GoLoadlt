package expression

import (
	"testing"

	"github.com/rinjold/go-etl-studio/internal/etl/contracts"
)

func TestEvalBool(t *testing.T) {
	row := contracts.DataRow{
		"amount":  120,
		"country": "FR",
		"qty":     0,
	}

	cases := []struct {
		expr string
		want bool
	}{
		{"amount > 100", true},
		{"amount < 100", false},
		{"amount >= 120", true},
		{"amount <= 119", false},
		{"amount == 120", true},
		{"amount != 0", true},
		{"qty == 0", true},
		{"country == 'FR'", true},
		{"country != 'DE'", true},
		{"country == 'DE'", false},
	}

	for _, tc := range cases {
		got, err := EvalBool(tc.expr, row)
		if err != nil {
			t.Fatalf("EvalBool(%q) error: %v", tc.expr, err)
		}
		if got != tc.want {
			t.Errorf("EvalBool(%q) = %v, want %v", tc.expr, got, tc.want)
		}
	}
}

func TestEvalValue(t *testing.T) {
	row := contracts.DataRow{
		"amount": float64(100),
		"tax":    float64(20),
		"country": "FR",
	}

	cases := []struct {
		expr string
		want any
	}{
		{"amount * 1.2", float64(120)},
		{"amount + tax", float64(120)},
		{"amount - tax", float64(80)},
		{"amount / tax", float64(5)},
		{"country", "FR"},
		{"'EU'", "EU"},
		{"42", float64(42)},
	}

	for _, tc := range cases {
		got, err := EvalValue(tc.expr, row)
		if err != nil {
			t.Fatalf("EvalValue(%q) error: %v", tc.expr, err)
		}
		if got != tc.want {
			t.Errorf("EvalValue(%q) = %#v, want %#v", tc.expr, got, tc.want)
		}
	}
}
