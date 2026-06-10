package interpreter

import (
	"fmt"

	"able/interpreter-go/pkg/ast"
)

type namedStructLiteralPlan struct {
	definition *ast.StructDefinition
	fieldOrder []int
}

func buildNamedStructLiteralPlan(lit *ast.StructLiteral, def *ast.StructDefinition) (namedStructLiteralPlan, error) {
	if lit == nil || def == nil || lit.StructType == nil || lit.StructType.Name == "" {
		return namedStructLiteralPlan{}, fmt.Errorf("nil named struct literal plan")
	}
	fieldOrder := make([]int, len(lit.Fields))
	seen := make([]bool, len(def.Fields))
	for idx, field := range lit.Fields {
		if field == nil || field.Name == nil || field.Name.Name == "" {
			return namedStructLiteralPlan{}, fmt.Errorf("named struct literal requires field names")
		}
		fieldIndex, ok := namedStructFieldIndex(def, field.Name.Name)
		if !ok {
			return namedStructLiteralPlan{}, fmt.Errorf("Unknown field '%s' for struct '%s'", field.Name.Name, lit.StructType.Name)
		}
		if seen[fieldIndex] {
			return namedStructLiteralPlan{}, fmt.Errorf("Duplicate field '%s' for struct '%s'", field.Name.Name, lit.StructType.Name)
		}
		seen[fieldIndex] = true
		fieldOrder[idx] = fieldIndex
	}
	for idx, defField := range def.Fields {
		if defField == nil || defField.Name == nil {
			continue
		}
		if !seen[idx] {
			return namedStructLiteralPlan{}, fmt.Errorf("Missing field '%s' for struct '%s'", defField.Name.Name, lit.StructType.Name)
		}
	}
	return namedStructLiteralPlan{
		definition: def,
		fieldOrder: fieldOrder,
	}, nil
}

func (i *Interpreter) namedStructLiteralPlanCached(lit *ast.StructLiteral, def *ast.StructDefinition) (namedStructLiteralPlan, error) {
	if lit == nil || def == nil {
		return buildNamedStructLiteralPlan(lit, def)
	}
	if i == nil {
		return buildNamedStructLiteralPlan(lit, def)
	}
	if i.envSingleThread {
		if plan, ok := i.namedStructLiteralPlanCache[lit]; ok && plan.definition == def {
			return plan, nil
		}
		plan, err := buildNamedStructLiteralPlan(lit, def)
		if err != nil {
			return namedStructLiteralPlan{}, err
		}
		i.namedStructLiteralPlanCache[lit] = plan
		return plan, nil
	}
	i.namedStructLiteralPlanCacheMu.RLock()
	plan, ok := i.namedStructLiteralPlanCache[lit]
	i.namedStructLiteralPlanCacheMu.RUnlock()
	if ok && plan.definition == def {
		return plan, nil
	}
	plan, err := buildNamedStructLiteralPlan(lit, def)
	if err != nil {
		return namedStructLiteralPlan{}, err
	}
	i.namedStructLiteralPlanCacheMu.Lock()
	if existing, ok := i.namedStructLiteralPlanCache[lit]; ok && existing.definition == def {
		i.namedStructLiteralPlanCacheMu.Unlock()
		return existing, nil
	}
	if i.namedStructLiteralPlanCache == nil {
		i.namedStructLiteralPlanCache = make(map[*ast.StructLiteral]namedStructLiteralPlan)
	}
	i.namedStructLiteralPlanCache[lit] = plan
	i.namedStructLiteralPlanCacheMu.Unlock()
	return plan, nil
}
