package interpreter

import "able/interpreter-go/pkg/ast"

type namedStructPatternPlanCacheKey struct {
	pattern    *ast.StructPattern
	definition *ast.StructDefinition
}

type namedStructPatternPlan struct {
	definition *ast.StructDefinition
	fieldOrder []int
}

func buildNamedStructPatternPlan(pattern *ast.StructPattern, def *ast.StructDefinition) (namedStructPatternPlan, bool) {
	if pattern == nil || def == nil || pattern.IsPositional {
		return namedStructPatternPlan{}, false
	}
	fieldOrder := make([]int, len(pattern.Fields))
	for idx, field := range pattern.Fields {
		if field == nil || field.FieldName == nil || field.FieldName.Name == "" {
			return namedStructPatternPlan{}, false
		}
		fieldIndex, ok := namedStructFieldIndex(def, field.FieldName.Name)
		if !ok {
			return namedStructPatternPlan{}, false
		}
		fieldOrder[idx] = fieldIndex
	}
	return namedStructPatternPlan{
		definition: def,
		fieldOrder: fieldOrder,
	}, true
}

func (i *Interpreter) namedStructPatternPlanCached(pattern *ast.StructPattern, def *ast.StructDefinition) (namedStructPatternPlan, bool) {
	if pattern == nil || def == nil {
		return buildNamedStructPatternPlan(pattern, def)
	}
	key := namedStructPatternPlanCacheKey{
		pattern:    pattern,
		definition: def,
	}
	if i == nil {
		return buildNamedStructPatternPlan(pattern, def)
	}
	if i.envSingleThread {
		if plan, ok := i.namedStructPatternPlanCache[key]; ok {
			return plan, true
		}
		plan, ok := buildNamedStructPatternPlan(pattern, def)
		if !ok {
			return namedStructPatternPlan{}, false
		}
		i.namedStructPatternPlanCache[key] = plan
		return plan, true
	}
	i.namedStructPatternPlanCacheMu.RLock()
	plan, ok := i.namedStructPatternPlanCache[key]
	i.namedStructPatternPlanCacheMu.RUnlock()
	if ok {
		return plan, true
	}
	plan, ok = buildNamedStructPatternPlan(pattern, def)
	if !ok {
		return namedStructPatternPlan{}, false
	}
	i.namedStructPatternPlanCacheMu.Lock()
	if existing, ok := i.namedStructPatternPlanCache[key]; ok {
		i.namedStructPatternPlanCacheMu.Unlock()
		return existing, true
	}
	if i.namedStructPatternPlanCache == nil {
		i.namedStructPatternPlanCache = make(map[namedStructPatternPlanCacheKey]namedStructPatternPlan)
	}
	i.namedStructPatternPlanCache[key] = plan
	i.namedStructPatternPlanCacheMu.Unlock()
	return plan, true
}
