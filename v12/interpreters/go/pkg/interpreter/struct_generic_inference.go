package interpreter

import (
	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

type structGenericInferenceField struct {
	fieldIndex int
	fieldName  string
	fieldType  ast.TypeExpression
}

type structGenericInferencePlan struct {
	expectedCount int
	namesByIndex  []string
	genericNames  map[string]struct{}
	genericIndex  map[string]int
	fields        []structGenericInferenceField
}

func buildStructGenericInferencePlan(def *ast.StructDefinition) *structGenericInferencePlan {
	if def == nil {
		return nil
	}
	genericNames := genericNameSet(def.GenericParams)
	fields := make([]structGenericInferenceField, 0, len(def.Fields))
	for idx, field := range def.Fields {
		if field == nil || field.FieldType == nil {
			continue
		}
		if !typeExpressionUsesGenerics(field.FieldType, genericNames) {
			continue
		}
		name := ""
		if field.Name != nil {
			name = field.Name.Name
		}
		fields = append(fields, structGenericInferenceField{
			fieldIndex: idx,
			fieldName:  name,
			fieldType:  field.FieldType,
		})
	}
	namesByIndex := buildGenericParamNamesByIndex(def.GenericParams)
	return &structGenericInferencePlan{
		expectedCount: len(def.GenericParams),
		namesByIndex:  namesByIndex,
		genericNames:  genericNames,
		genericIndex:  buildGenericParamIndex(namesByIndex),
		fields:        fields,
	}
}

func (i *Interpreter) structGenericInferencePlan(def *ast.StructDefinition) *structGenericInferencePlan {
	if i == nil || def == nil {
		return buildStructGenericInferencePlan(def)
	}
	if i.envSingleThread {
		if plan, ok := i.structGenericInferencePlanCache[def]; ok {
			return plan
		}
		plan := buildStructGenericInferencePlan(def)
		i.structGenericInferencePlanCache[def] = plan
		return plan
	}
	i.structGenericInferencePlanCacheMu.RLock()
	plan, ok := i.structGenericInferencePlanCache[def]
	i.structGenericInferencePlanCacheMu.RUnlock()
	if ok {
		return plan
	}
	plan = buildStructGenericInferencePlan(def)
	i.structGenericInferencePlanCacheMu.Lock()
	if existing, ok := i.structGenericInferencePlanCache[def]; ok {
		i.structGenericInferencePlanCacheMu.Unlock()
		return existing
	}
	i.structGenericInferencePlanCache[def] = plan
	i.structGenericInferencePlanCacheMu.Unlock()
	return plan
}

func structTypeArgsNeedInference(plan *structGenericInferencePlan, typeArgs []ast.TypeExpression) bool {
	if plan == nil {
		return len(typeArgs) > 0
	}
	if len(typeArgs) != plan.expectedCount {
		return true
	}
	for _, arg := range typeArgs {
		if arg == nil {
			return true
		}
		if _, ok := arg.(*ast.WildcardTypeExpression); ok {
			return true
		}
		simple, ok := arg.(*ast.SimpleTypeExpression)
		if !ok || simple == nil || simple.Name == nil || simple.Name.Name == "" {
			continue
		}
		if _, isGeneric := plan.genericNames[simple.Name.Name]; isGeneric {
			return true
		}
	}
	return false
}

func structTypeArgsConcreteForDefinition(def *ast.StructDefinition, typeArgs []ast.TypeExpression) bool {
	if def == nil || len(typeArgs) != len(def.GenericParams) {
		return false
	}
	for _, arg := range typeArgs {
		if arg == nil {
			return false
		}
		if _, ok := arg.(*ast.WildcardTypeExpression); ok {
			return false
		}
		simple, ok := arg.(*ast.SimpleTypeExpression)
		if !ok || simple == nil || simple.Name == nil || simple.Name.Name == "" {
			continue
		}
		for _, param := range def.GenericParams {
			if param != nil && param.Name != nil && simple.Name.Name == param.Name.Name {
				return false
			}
		}
	}
	return true
}

func (i *Interpreter) inferStructTypeArgumentsWithSeen(def *ast.StructDefinition, named map[string]runtime.Value, positional []runtime.Value, seen map[*runtime.StructInstanceValue]struct{}) []ast.TypeExpression {
	plan := i.structGenericInferencePlan(def)
	return i.inferStructTypeArgumentsFromPlanWithSeen(plan, named, positional, seen)
}

func (i *Interpreter) inferStructTypeArgumentsFromPlanWithSeen(plan *structGenericInferencePlan, named map[string]runtime.Value, positional []runtime.Value, seen map[*runtime.StructInstanceValue]struct{}) []ast.TypeExpression {
	if plan == nil || plan.expectedCount == 0 {
		return nil
	}
	var inline [3]ast.TypeExpression
	bindings := inlineBindingsSlice(inline[:], plan.expectedCount)
	for _, field := range plan.fields {
		var (
			val runtime.Value
			ok  bool
		)
		if named != nil && field.fieldName != "" {
			val, ok = named[field.fieldName]
		}
		if !ok && field.fieldIndex < len(positional) {
			val, ok = positional[field.fieldIndex], true
		}
		if !ok {
			continue
		}
		actual := i.typeExpressionFromValueWithSeen(val, seen)
		if actual == nil {
			continue
		}
		matchTypeExpressionTemplateIndexed(field.fieldType, actual, plan.genericIndex, bindings)
		if indexedTypeExpressionBindingsComplete(bindings) {
			break
		}
	}
	return i.cachedTypeArgumentsFromIndexedBindings(bindings)
}
