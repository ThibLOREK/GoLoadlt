package expression

import (
	"testing"

	"github.com/rinjold/go-etl-studio/internal/etl/contracts"
)

func TestEvalBool(t *testing.T) {
	row := contracts.DataRow{
		"amount": 120,
		"country": "FR",
	}

	cases := []struct {
		expr string
		want bool
	}{
		{"amount > 100", true},
		{"amount < 100", false},
		{"amount >= 120", true},
		{"country == 'FR'", true},
		{"country != 'DE'", true},
	}

	for _, tc := range cases {
		got, err := EvalBool(tc.expr, row)
		if err != nil {
			t.Fatalf("EvalBool(%q) error: %v", tc.expr, err)
		}
		if got != tc.want {
			t.Fatalf("EvalBool(%q) = %v, want %v", tc.expr, got, tc.want)
		}
	}
}

func TestEvalValue(t *testing.T) {
	row := contracts.DataRow{
		"amount": 100,
		"tax": 20,
		"country": "FR",
	}

	cases := []struct {
		expr string
		want any
	}{
		{"amount * 1.2", 120.0},
		{"amount + tax", 120.0},
		{"country", "FR"},
		{"'EU'", "EU"},
	}

	for _, tc := range cases {
		got, err := EvalValue(tc.expr, row)
		if err != nil {
			t.Fatalf("EvalValue(%q) error: %v", tc.expr, err)
		}
		if got != tc.want {
			t.Fatalf("EvalValue(%q) = %#v, want %#v", tc.expr, got, tc.want)
		}
	}
}
