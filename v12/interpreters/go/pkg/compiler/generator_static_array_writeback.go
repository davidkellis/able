package compiler

import (
	"fmt"

	"able/interpreter-go/pkg/ast"
)

func (g *generator) appendRecoveredStaticArrayWriteback(ctx *compileContext, objNode ast.Expression, mutatedExpr string, mutatedType string) ([]string, bool) {
	if g == nil || ctx == nil || objNode == nil || mutatedExpr == "" || mutatedType == "" || !g.isStaticArrayType(mutatedType) {
		return nil, false
	}
	ident, ok := objNode.(*ast.Identifier)
	if !ok || ident == nil || ident.Name == "" {
		return nil, false
	}
	binding, ok := ctx.lookup(ident.Name)
	if !ok || binding.GoName == "" || binding.GoType == "" || binding.GoType == mutatedType || !g.isStaticArrayType(binding.GoType) {
		return nil, false
	}
	valueLines, valueExpr, ok := g.lowerRuntimeValue(ctx, mutatedExpr, mutatedType)
	if !ok {
		return nil, false
	}
	convLines, converted, ok := g.lowerExpectRuntimeValue(ctx, valueExpr, binding.GoType)
	if !ok {
		return nil, false
	}
	lines := append([]string{}, valueLines...)
	lines = append(lines, convLines...)
	lines = append(lines, fmt.Sprintf("%s = %s", binding.GoName, converted))
	return lines, true
}
