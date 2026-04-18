package transformers

import (
	"context"
	"fmt"

	"github.com/rinjold/go-etl-studio/internal/etl/contracts"
)

type Operator string

const (
	OpEq  Operator = "eq"
	OpNeq Operator = "neq"
	OpGt  Operator = "gt"
	OpLt  Operator = "lt"
)

type FilterRule struct {
	Column   string
	Operator Operator
	Value    string
}

// Filter conserve uniquement les enregistrements qui satisfont toutes les règles.
type Filter struct {
	Rules []FilterRule
}

func (f Filter) Transform(ctx context.Context, in []contracts.Record) ([]contracts.Record, error) {
	out := make([]contracts.Record, 0, len(in))
	for _, rec := range in {
		match, err := f.matches(rec)
		if err != nil {
			return nil, err
		}
		if match {
			out = append(out, rec)
		}
	}
	return out, nil
}

func (f Filter) matches(rec contracts.Record) (bool, error) {
	for _, rule := range f.Rules {
		val, ok := rec[rule.Column]
		if !ok {
			return false, nil
		}
		strVal := fmt.Sprintf("%v", val)
		switch rule.Operator {
		case OpEq:
			if strVal != rule.Value {
				return false, nil
			}
		case OpNeq:
			if strVal == rule.Value {
				return false, nil
			}
		case OpGt:
			if strVal <= rule.Value {
				return false, nil
			}
		case OpLt:
			if strVal >= rule.Value {
				return false, nil
			}
		default:
			return false, fmt.Errorf("unknown operator: %s", rule.Operator)
		}
	}
	return true, nil
}
