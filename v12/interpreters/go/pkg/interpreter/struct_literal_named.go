package interpreter

import (
	"fmt"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func simpleNamedStructLiteralDefinitionEligible(def *ast.StructDefinition, name string) bool {
	if def == nil || name == "" {
		return false
	}
	if def.Kind != ast.StructKindNamed {
		return false
	}
	return true
}

func namedStructFieldIndex(def *ast.StructDefinition, name string) (int, bool) {
	return structDefinitionNamedFieldIndex(def, name)
}

func namedStructLiteralFieldOrder(lit *ast.StructLiteral, def *ast.StructDefinition) ([]int, error) {
	plan, err := buildNamedStructLiteralPlan(lit, def)
	if err != nil {
		return nil, err
	}
	return plan.fieldOrder, nil
}

func simpleNamedStructLiteralSyntacticEligible(lit *ast.StructLiteral) bool {
	if lit == nil || lit.StructType == nil || lit.StructType.Name == "" {
		return false
	}
	if lit.IsPositional || len(lit.FunctionalUpdateSources) != 0 {
		return false
	}
	for _, field := range lit.Fields {
		value, ok := simpleNamedStructLiteralFieldValue(field)
		if !ok {
			return false
		}
		if field.Name == nil || field.Name.Name == "" {
			ident, ok := value.(*ast.Identifier)
			if !field.IsShorthand || !ok || ident == nil || ident.Name == "" {
				return false
			}
		}
	}
	return true
}

// simpleNamedStructLiteralFieldValue returns the expression represented by a
// field initializer. Parser-produced shorthand fields keep the identifier in
// Name and intentionally leave Value nil; synthetic ASTs may keep it in Value.
func simpleNamedStructLiteralFieldValue(field *ast.StructFieldInitializer) (ast.Expression, bool) {
	if field == nil {
		return nil, false
	}
	if field.Value != nil {
		return field.Value, true
	}
	if field.IsShorthand && field.Name != nil && field.Name.Name != "" {
		return field.Name, true
	}
	return nil, false
}

func simpleNamedStructLiteralEligibleForDefinition(lit *ast.StructLiteral, def *ast.StructDefinition) bool {
	if def == nil || !simpleNamedStructLiteralSyntacticEligible(lit) {
		return false
	}
	return simpleNamedStructLiteralDefinitionEligible(def, lit.StructType.Name)
}

func (i *Interpreter) singletonStructLiteralValueIfPossible(structDefVal *runtime.StructDefinitionValue, structName string, explicitTypeArgs []ast.TypeExpression) (runtime.Value, bool, error) {
	if i == nil || structDefVal == nil || structDefVal.Node == nil || !isSingletonStructDef(structDefVal.Node) {
		return nil, false, nil
	}
	typeArgs, err := i.resolveStructTypeArguments(structDefVal.Node, explicitTypeArgs, nil, nil, nil)
	if err != nil {
		return nil, true, err
	}
	if err := i.enforceStructConstraints(structDefVal.Node, typeArgs, structName); err != nil {
		return nil, true, err
	}
	return structDefVal, true, nil
}

func namedStructFieldMapFromPositional(def *ast.StructDefinition, values []runtime.Value) map[string]runtime.Value {
	if def == nil || len(values) == 0 {
		return nil
	}
	fields := make(map[string]runtime.Value, len(values))
	for idx, field := range def.Fields {
		if field == nil || field.Name == nil || idx >= len(values) {
			continue
		}
		fields[field.Name.Name] = values[idx]
	}
	return fields
}

func (i *Interpreter) evaluateSimpleNamedStructLiteralIfPossible(lit *ast.StructLiteral, env *runtime.Environment, structDefVal *runtime.StructDefinitionValue, explicitTypeArgs []ast.TypeExpression) (runtime.Value, bool, error) {
	if i == nil || lit == nil || env == nil || structDefVal == nil || structDefVal.Node == nil {
		return nil, false, nil
	}
	structDef := structDefVal.Node
	if !simpleNamedStructLiteralEligibleForDefinition(lit, structDef) {
		return nil, false, nil
	}
	plan, err := i.namedStructLiteralPlanCached(lit, structDef)
	if err != nil {
		return nil, false, err
	}
	if singletonValue, ok, err := i.singletonStructLiteralValueIfPossible(structDefVal, lit.StructType.Name, explicitTypeArgs); ok {
		return singletonValue, true, err
	}
	inst, values := runtime.NewStructInstancePositionalSized(structDefVal, len(structDef.Fields), nil)
	for idx, field := range lit.Fields {
		value, ok := simpleNamedStructLiteralFieldValue(field)
		if !ok {
			return nil, false, fmt.Errorf("named struct literal field has no value")
		}
		val, err := i.evaluateExpression(value, env)
		if err != nil {
			return nil, true, err
		}
		fieldIndex := plan.fieldOrder[idx]
		values[fieldIndex] = val
	}
	typeArgs, err := i.resolveStructTypeArguments(structDef, explicitTypeArgs, nil, nil, values)
	if err != nil {
		return nil, true, err
	}
	if err := i.enforceStructConstraints(structDef, typeArgs, lit.StructType.Name); err != nil {
		return nil, true, err
	}
	if lit.StructType.Name == "Array" {
		array, err := i.arrayValueFromStructFieldValues(structDef.Fields, values)
		return array, true, err
	}
	inst.TypeArguments = typeArgs
	return inst, true, nil
}
