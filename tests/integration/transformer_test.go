package integration

import (
	"context"
	"testing"

	"github.com/rinjold/go-etl-studio/internal/etl/contracts"
	"github.com/rinjold/go-etl-studio/internal/etl/transformers"
)

func TestMapper(t *testing.T) {
	mapper := transformers.Mapper{Mapping: map[string]string{"prenom": "first_name"}}
	in := []contracts.Record{{"prenom": "Alice", "age": "30"}}
	out, err := mapper.Transform(context.Background(), in)
	if err != nil {
		t.Fatal(err)
	}
	if out[0]["first_name"] != "Alice" {
		t.Errorf("expected first_name=Alice, got %v", out[0]["first_name"])
	}
	if _, ok := out[0]["prenom"]; ok {
		t.Error("old key 'prenom' should have been removed")
	}
}

func TestFilter_Eq(t *testing.T) {
	filter := transformers.Filter{Rules: []transformers.FilterRule{
		{Column: "status", Operator: transformers.OpEq, Value: "active"},
	}}
	in := []contracts.Record{
		{"status": "active", "name": "Alice"},
		{"status": "inactive", "name": "Bob"},
	}
	out, err := filter.Transform(context.Background(), in)
	if err != nil {
		t.Fatal(err)
	}
	if len(out) != 1 || out[0]["name"] != "Alice" {
		t.Errorf("expected 1 active record, got %d", len(out))
	}
}

func TestCaster_Int(t *testing.T) {
	caster := transformers.Caster{Rules: []transformers.CastRule{
		{Column: "age", CastTo: transformers.CastInt},
	}}
	in := []contracts.Record{{"age": "30"}}
	out, err := caster.Transform(context.Background(), in)
	if err != nil {
		t.Fatal(err)
	}
	val, ok := out[0]["age"].(int64)
	if !ok || val != 30 {
		t.Errorf("expected int64(30), got %T %v", out[0]["age"], out[0]["age"])
	}
}

func TestChain(t *testing.T) {
	chain := transformers.Chain{Steps: []contracts.Transformer{
		transformers.Mapper{Mapping: map[string]string{"prenom": "first_name"}},
		transformers.Filter{Rules: []transformers.FilterRule{
			{Column: "first_name", Operator: transformers.OpEq, Value: "Alice"},
		}},
	}}
	in := []contracts.Record{
		{"prenom": "Alice"},
		{"prenom": "Bob"},
	}
	out, err := chain.Transform(context.Background(), in)
	if err != nil {
		t.Fatal(err)
	}
	if len(out) != 1 || out[0]["first_name"] != "Alice" {
		t.Errorf("expected 1 record with first_name=Alice, got %+v", out)
	}
}
