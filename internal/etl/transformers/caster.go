package transformers

import (
	"context"
	"fmt"
	"strconv"

	"github.com/rinjold/go-etl-studio/internal/etl/contracts"
)

type CastType string

const (
	CastInt     CastType = "int"
	CastFloat   CastType = "float"
	CastBool    CastType = "bool"
	CastString  CastType = "string"
)

type CastRule struct {
	Column   string
	CastTo   CastType
}

// Caster convertit les valeurs de colonnes vers les types Go natifs.
type Caster struct {
	Rules []CastRule
}

func (c Caster) Transform(ctx context.Context, in []contracts.Record) ([]contracts.Record, error) {
	out := make([]contracts.Record, 0, len(in))
	for _, rec := range in {
		newRec := make(contracts.Record, len(rec))
		for k, v := range rec {
			newRec[k] = v
		}
		for _, rule := range c.Rules {
			raw, ok := newRec[rule.Column]
			if !ok {
				continue
			}
			casted, err := castValue(fmt.Sprintf("%v", raw), rule.CastTo)
			if err != nil {
				return nil, fmt.Errorf("cast column %q to %s: %w", rule.Column, rule.CastTo, err)
			}
			newRec[rule.Column] = casted
		}
		out = append(out, newRec)
	}
	return out, nil
}

func castValue(raw string, t CastType) (any, error) {
	switch t {
	case CastInt:
		return strconv.ParseInt(raw, 10, 64)
	case CastFloat:
		return strconv.ParseFloat(raw, 64)
	case CastBool:
		return strconv.ParseBool(raw)
	case CastString:
		return raw, nil
	default:
		return nil, fmt.Errorf("unknown cast type: %s", t)
	}
}
