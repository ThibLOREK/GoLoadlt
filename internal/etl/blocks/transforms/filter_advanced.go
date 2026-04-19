package transforms

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/rinjold/go-etl-studio/internal/etl/blocks"
	"github.com/rinjold/go-etl-studio/internal/etl/contracts"
)

func init() {
	blocks.Register("transform.filter_advanced", func() contracts.Block { return &FilterAdvanced{} })
}

// FilterAdvanced route les lignes vers deux sorties (true/false) selon une condition.
//
// Paramètres :
//   - field      : nom de la colonne à évaluer (obligatoire)
//   - operator   : eq | neq | gt | gte | lt | lte | contains | not_contains |
//                  starts_with | ends_with | is_null | is_not_null | is_true | is_false
//   - value      : valeur de comparaison (ignorée pour is_null/is_not_null/is_true/is_false)
//   - value_type : string | number | bool (défaut: string)
//
// Sorties :
//   - Outputs[0] : lignes qui satisfont la condition (branche true)
//   - Outputs[1] : lignes qui ne satisfont pas (branche false) — optionnel
type FilterAdvanced struct{}

func (b *FilterAdvanced) Type() string { return "transform.filter_advanced" }

func (b *FilterAdvanced) Run(bctx *contracts.BlockContext) error {
	if len(bctx.Inputs) == 0 {
		return fmt.Errorf("transform.filter_advanced: aucun port d'entrée")
	}
	field := bctx.Params["field"]
	operator := bctx.Params["operator"]
	value := bctx.Params["value"]
	valueType := bctx.Params["value_type"]
	if valueType == "" {
		valueType = "string"
	}
	if field == "" || operator == "" {
		return fmt.Errorf("transform.filter_advanced: paramètres 'field' et 'operator' obligatoires")
	}
	if len(bctx.Outputs) == 0 {
		return fmt.Errorf("transform.filter_advanced: au moins un port de sortie requis")
	}

	trueCh := bctx.Outputs[0].Ch
	var falseCh chan contracts.DataRow
	if len(bctx.Outputs) > 1 {
		falseCh = bctx.Outputs[1].Ch
	}

	closeAll := func() {
		for _, out := range bctx.Outputs {
			close(out.Ch)
		}
	}

	in := bctx.Inputs[0]
	for {
		select {
		case <-bctx.Ctx.Done():
			closeAll()
			return bctx.Ctx.Err()
		case row, ok := <-in.Ch:
			if !ok {
				closeAll()
				return nil
			}
			match, err := evalCondition(row, field, operator, value, valueType)
			if err != nil {
				// CORRECTION : fermer les sorties avant de retourner l'erreur
				// pour éviter de bloquer les consommateurs en attente.
				closeAll()
				return fmt.Errorf("transform.filter_advanced: %w", err)
			}
			if match {
				select {
				case trueCh <- row:
				case <-bctx.Ctx.Done():
					closeAll()
					return bctx.Ctx.Err()
				}
			} else if falseCh != nil {
				select {
				case falseCh <- row:
				case <-bctx.Ctx.Done():
					closeAll()
					return bctx.Ctx.Err()
				}
			}
			// Si falseCh == nil et match == false : la ligne est silencieusement
			// droppée (comportement attendu quand la branche false n'est pas câblée).
		}
	}
}

func evalCondition(row contracts.DataRow, field, operator, value, valueType string) (bool, error) {
	raw, exists := row[field]

	switch operator {
	case "is_null":
		return !exists || raw == nil, nil
	case "is_not_null":
		return exists && raw != nil, nil
	}

	if !exists || raw == nil {
		// Champ absent ou null : la condition est fausse pour tous les autres opérateurs.
		return false, nil
	}

	switch valueType {
	case "number":
		rowNum, err := toFloat(raw)
		if err != nil {
			return false, fmt.Errorf("champ '%s' non numérique: %v", field, raw)
		}
		cmpNum, err := strconv.ParseFloat(strings.TrimSpace(value), 64)
		if err != nil {
			return false, fmt.Errorf("valeur de comparaison '%s' non numérique", value)
		}
		switch operator {
		case "eq":
			return rowNum == cmpNum, nil
		case "neq":
			return rowNum != cmpNum, nil
		case "gt":
			return rowNum > cmpNum, nil
		case "gte":
			return rowNum >= cmpNum, nil
		case "lt":
			return rowNum < cmpNum, nil
		case "lte":
			return rowNum <= cmpNum, nil
		default:
			return false, fmt.Errorf("opérateur '%s' non supporté pour type number", operator)
		}

	case "bool":
		rowBool, err := toBool(raw)
		if err != nil {
			return false, fmt.Errorf("champ '%s' non booléen: %v", field, raw)
		}
		switch operator {
		case "is_true", "eq":
			return rowBool, nil
		case "is_false", "neq":
			return !rowBool, nil
		default:
			return false, fmt.Errorf("opérateur '%s' non supporté pour type bool", operator)
		}

	default: // string
		rowStr := fmt.Sprintf("%v", raw)
		switch operator {
		case "eq":
			return rowStr == value, nil
		case "neq":
			return rowStr != value, nil
		case "contains":
			return strings.Contains(rowStr, value), nil
		case "not_contains":
			return !strings.Contains(rowStr, value), nil
		case "starts_with":
			return strings.HasPrefix(rowStr, value), nil
		case "ends_with":
			return strings.HasSuffix(rowStr, value), nil
		case "gt":
			return rowStr > value, nil
		case "lt":
			return rowStr < value, nil
		default:
			return false, fmt.Errorf("opérateur '%s' non supporté pour type string", operator)
		}
	}
}

func toFloat(v any) (float64, error) {
	switch val := v.(type) {
	case float64:
		return val, nil
	case float32:
		return float64(val), nil
	case int:
		return float64(val), nil
	case int64:
		return float64(val), nil
	case string:
		return strconv.ParseFloat(strings.TrimSpace(val), 64)
	default:
		return 0, fmt.Errorf("type inconnu pour conversion numérique: %T", v)
	}
}

func toBool(v any) (bool, error) {
	switch val := v.(type) {
	case bool:
		return val, nil
	case string:
		return strconv.ParseBool(strings.TrimSpace(val))
	default:
		f, err := toFloat(v)
		if err != nil {
			return false, fmt.Errorf("type inconnu pour conversion booléenne: %T", v)
		}
		return f != 0, nil
	}
}
