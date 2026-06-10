package interpreter

import (
	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func structHasGenericConstraints(def *ast.StructDefinition) bool {
	if def == nil || len(def.GenericParams) == 0 {
		return false
	}
	return hasAnyGenericConstraints(def.GenericParams, def.WhereClause)
}

func (i *Interpreter) resolvedStructInstanceTypeArguments(inst *runtime.StructInstanceValue) []ast.TypeExpression {
	return i.resolvedStructInstanceTypeArgumentsWithSeen(inst, nil)
}
