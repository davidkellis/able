package compiler

import (
	"fmt"

	"able/interpreter-go/pkg/ast"
)

func (g *generator) recoverDispatchBinding(ctx *compileContext, binding paramInfo) paramInfo {
	if g == nil {
		return binding
	}
	if binding.GoType != "" && binding.GoType != "runtime.Value" && binding.GoType != "any" {
		return binding
	}
	if binding.TypeExpr == nil {
		return binding
	}
	if recovered, ok := g.joinCarrierTypeFromTypeExpr(ctx, binding.TypeExpr); ok && recovered != "" && recovered != "runtime.Value" && recovered != "any" && !g.isVoidType(recovered) {
		binding.GoType = recovered
	}
	return binding
}

func (g *generator) recoverDispatchCarrierType(ctx *compileContext, expr ast.Expression, goType string) (string, bool) {
	if g == nil || expr == nil {
		return "", false
	}
	if goType != "" && goType != "runtime.Value" && goType != "any" && !g.isVoidType(goType) {
		return goType, true
	}
	inferred, ok := g.inferExpressionTypeExpr(ctx, expr, goType)
	if !ok || inferred == nil {
		return "", false
	}
	recovered, ok := g.joinCarrierTypeFromTypeExpr(ctx, inferred)
	if !ok || recovered == "" || recovered == "runtime.Value" || recovered == "any" || g.isVoidType(recovered) {
		return "", false
	}
	return recovered, true
}

func (g *generator) recoverDispatchExpr(ctx *compileContext, original ast.Expression, compiledExpr string, compiledType string) ([]string, string, string, bool) {
	if g == nil || ctx == nil || original == nil || compiledExpr == "" || compiledType == "" {
		return nil, compiledExpr, compiledType, false
	}
	recoveredType, ok := g.recoverDispatchCarrierType(ctx, original, compiledType)
	if !ok || recoveredType == compiledType {
		return nil, compiledExpr, compiledType, false
	}
	switch compiledType {
	case "runtime.Value":
		lines, converted, ok := g.lowerExpectRuntimeValue(ctx, compiledExpr, recoveredType)
		if !ok {
			return nil, compiledExpr, compiledType, false
		}
		return lines, converted, recoveredType, true
	case "any":
		valueTemp := ctx.newTemp()
		lines := []string{fmt.Sprintf("%s := __able_any_to_value(%s)", valueTemp, compiledExpr)}
		convLines, converted, ok := g.lowerExpectRuntimeValue(ctx, valueTemp, recoveredType)
		if !ok {
			return nil, compiledExpr, compiledType, false
		}
		lines = append(lines, convLines...)
		return lines, converted, recoveredType, true
	default:
		return nil, compiledExpr, compiledType, false
	}
}
