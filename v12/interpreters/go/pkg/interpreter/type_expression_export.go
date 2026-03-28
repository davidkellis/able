package interpreter

import (
	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

// TypeExpressionFromValue exposes the internal type expression inference for compiler helpers.
func (i *Interpreter) TypeExpressionFromValue(value runtime.Value) ast.TypeExpression {
	if i == nil {
		return nil
	}
	return i.typeExpressionFromValue(value)
}

// ExpandTypeAliases exposes alias expansion for compiler helpers.
func (i *Interpreter) ExpandTypeAliases(expr ast.TypeExpression) ast.TypeExpression {
	if i == nil {
		return expr
	}
	return expandTypeAliases(expr, i.typeAliases, nil)
}

// LookupUnionDefinition exposes named union lookup for compiler bridge helpers.
func (i *Interpreter) LookupUnionDefinition(name string) (*runtime.UnionDefinitionValue, bool) {
	if i == nil || name == "" {
		return nil, false
	}
	def, ok := i.unionDefinitions[name]
	if !ok || def == nil {
		return nil, false
	}
	return def, true
}

// EnsureTypeSatisfiesInterface exposes interface constraint checks for compiler helpers.
func (i *Interpreter) EnsureTypeSatisfiesInterface(subject ast.TypeExpression, ifaceExpr ast.TypeExpression, context string) error {
	if i == nil {
		return nil
	}
	info, ok := parseTypeExpression(subject)
	if !ok {
		return nil
	}
	return i.ensureTypeSatisfiesInterface(info, ifaceExpr, context, make(map[string]struct{}))
}

// IsKnownConstraintTypeName reports whether a name should be treated as known for constraint enforcement.
func (i *Interpreter) IsKnownConstraintTypeName(name string) bool {
	if i == nil {
		return false
	}
	if name == "" || name == "_" || name == "Self" {
		return false
	}
	if isPrimitiveTypeName(name) {
		return true
	}
	if _, ok := i.interfaces[name]; ok {
		return true
	}
	if _, ok := i.unionDefinitions[name]; ok {
		return true
	}
	if _, ok := i.typeAliases[name]; ok {
		return true
	}
	if _, ok := i.lookupStructDefinition(name); ok {
		return true
	}
	return false
}
