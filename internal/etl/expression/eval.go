package expression

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/rinjold/go-etl-studio/internal/etl/contracts"
)

// EvalBool évalue une expression booléenne simple sur une ligne.
// Exemples supportés :
//   amount > 100
//   country == 'FR'
//   status != "archived"
//   qty >= 10
func EvalBool(expr string, row contracts.DataRow) (bool, error) {
	expr = strings.TrimSpace(expr)
	for _, op := range []string{"==", "!=", ">=", "<=", ">", "<"} {
		if idx := strings.Index(expr, op); idx >= 0 {
			left := strings.TrimSpace(expr[:idx])
			right := strings.TrimSpace(expr[idx+len(op):])
			return compare(row[left], right, op)
		}
	}
	return false, fmt.Errorf("expression booléenne non supportée: %s", expr)
}

// EvalValue évalue une expression simple retournant une valeur.
// Exemples supportés :
//   amount * 1.2
//   amount + tax
//   'FR'
//   status
func EvalValue(expr string, row contracts.DataRow) (any, error) {
	expr = strings.TrimSpace(expr)

	// Littéral string
	if s, ok := parseString(expr); ok {
		return s, nil
	}

	// Opérations binaires simples
	for _, op := range []string{"+", "-", "*", "/"} {
		if idx := strings.Index(expr, op); idx >= 0 {
			left := strings.TrimSpace(expr[:idx])
			right := strings.TrimSpace(expr[idx+1:])
			lv, err := resolveNumeric(left, row)
			if err != nil {
				return nil, err
			}
			rv, err := resolveNumeric(right, row)
			if err != nil {
				return nil, err
			}
			switch op {
			case "+":
				return lv + rv, nil
			case "-":
				return lv - rv, nil
			case "*":
				return lv * rv, nil
			case "/":
				if rv == 0 {
					return nil, fmt.Errorf("division par zéro")
				}
				return lv / rv, nil
			}
		}
	}

	// Colonne seule
	if v, ok := row[expr]; ok {
		return v, nil
	}

	// Littéral numérique
	if n, err := strconv.ParseFloat(expr, 64); err == nil {
		return n, nil
	}

	return nil, fmt.Errorf("expression non supportée: %s", expr)
}

func compare(left any, rightExpr, op string) (bool, error) {
	if rs, ok := parseString(rightExpr); ok {
		ls := fmt.Sprintf("%v", left)
		switch op {
		case "==":
			return ls == rs, nil
		case "!=":
			return ls != rs, nil
		default:
			return false, fmt.Errorf("opérateur %s non supporté pour string", op)
		}
	}

	lf, err := toFloat(left)
	if err != nil {
		return false, err
	}
	rf, err := strconv.ParseFloat(rightExpr, 64)
	if err != nil {
		return false, fmt.Errorf("valeur numérique invalide: %s", rightExpr)
	}

	switch op {
	case "==":
		return lf == rf, nil
	case "!=":
		return lf != rf, nil
	case ">":
		return lf > rf, nil
	case "<":
		return lf < rf, nil
	case ">=":
		return lf >= rf, nil
	case "<=":
		return lf <= rf, nil
	default:
		return false, fmt.Errorf("opérateur non supporté: %s", op)
	}
}

func resolveNumeric(token string, row contracts.DataRow) (float64, error) {
	if v, ok := row[token]; ok {
		return toFloat(v)
	}
	return strconv.ParseFloat(token, 64)
}

func toFloat(v any) (float64, error) {
	switch x := v.(type) {
	case int:
		return float64(x), nil
	case int32:
		return float64(x), nil
	case int64:
		return float64(x), nil
	case float32:
		return float64(x), nil
	case float64:
		return x, nil
	case string:
		return strconv.ParseFloat(x, 64)
	default:
		return strconv.ParseFloat(fmt.Sprintf("%v", v), 64)
	}
}

func parseString(s string) (string, bool) {
	if len(s) >= 2 {
		if (s[0] == '\'' && s[len(s)-1] == '\'') || (s[0] == '"' && s[len(s)-1] == '"') {
			return s[1 : len(s)-1], true
		}
	}
	return "", false
}
