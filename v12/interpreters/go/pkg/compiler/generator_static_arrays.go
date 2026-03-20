package compiler

import (
	"fmt"

	"able/interpreter-go/pkg/ast"
)

func (g *generator) inferMonoArraySpecForElementTypes(elementTypes []string) (*monoArraySpec, bool) {
	if g == nil || !g.monoArraysEnabled() || len(elementTypes) == 0 {
		return nil, false
	}
	var found *monoArraySpec
	for _, elemType := range elementTypes {
		spec, ok := g.monoArraySpecForElementGoType(elemType)
		if !ok || spec == nil {
			return nil, false
		}
		if found == nil {
			found = spec
			continue
		}
		if found.GoType != spec.GoType {
			return nil, false
		}
	}
	return found, found != nil
}

func (g *generator) staticArrayCloneLines(ctx *compileContext, arrayType string, valuesExpr string, capacityExpr string) ([]string, string, bool) {
	if ctx == nil || arrayType == "" || valuesExpr == "" {
		return nil, "", false
	}
	if capacityExpr == "" {
		capacityExpr = fmt.Sprintf("len(%s)", valuesExpr)
	}
	valuesTemp := ctx.newTemp()
	arrayTemp := ctx.newTemp()
	if spec, ok := g.monoArraySpecForGoType(arrayType); ok && spec != nil {
		lines := []string{
			fmt.Sprintf("%s := make([]%s, len(%s), %s)", valuesTemp, spec.ElemGoType, valuesExpr, capacityExpr),
			fmt.Sprintf("copy(%s, %s)", valuesTemp, valuesExpr),
			fmt.Sprintf("%s := &%s{Elements: %s}", arrayTemp, spec.GoName, valuesTemp),
			fmt.Sprintf("%s(%s)", spec.SyncHelper, arrayTemp),
		}
		return lines, arrayTemp, true
	}
	if g.isArrayStructType(arrayType) {
		lines := []string{
			fmt.Sprintf("%s := make([]runtime.Value, len(%s), %s)", valuesTemp, valuesExpr, capacityExpr),
			fmt.Sprintf("copy(%s, %s)", valuesTemp, valuesExpr),
			fmt.Sprintf("%s := &Array{Elements: %s}", arrayTemp, valuesTemp),
			fmt.Sprintf("__able_struct_Array_sync(%s)", arrayTemp),
		}
		return lines, arrayTemp, true
	}
	return nil, "", false
}

func (g *generator) inferStaticArrayTypeExpr(ctx *compileContext, expr ast.Expression, goType string) (ast.TypeExpression, bool) {
	if g == nil {
		return nil, false
	}
	if spec, ok := g.monoArraySpecForGoType(goType); ok && spec != nil {
		innerExpr, ok := g.typeExprForGoType(spec.ElemGoType)
		if !ok {
			return nil, false
		}
		return ast.Gen(ast.Ty("Array"), innerExpr), true
	}
	if !g.isArrayStructType(goType) {
		return nil, false
	}
	switch e := expr.(type) {
	case *ast.Identifier:
		if ctx == nil {
			return nil, false
		}
		if binding, ok := ctx.lookup(e.Name); ok && binding.TypeExpr != nil {
			return g.typeExprInContext(ctx, binding.TypeExpr), true
		}
	case *ast.FunctionCall:
		if inferred, ok := g.inferStaticCallResultTypeExpr(ctx, e); ok && inferred != nil {
			return inferred, true
		}
	case *ast.ArrayLiteral:
		return g.inferStaticArrayLiteralTypeExpr(ctx, e)
	}
	return nil, false
}

func (g *generator) inferStaticCallResultTypeExpr(ctx *compileContext, call *ast.FunctionCall) (ast.TypeExpression, bool) {
	if g == nil || ctx == nil || call == nil || call.Callee == nil {
		return nil, false
	}
	switch callee := call.Callee.(type) {
	case *ast.Identifier:
		if info, _, ok := g.resolveStaticCallable(ctx, callee.Name); ok && info != nil && info.Definition != nil && info.Definition.ReturnType != nil {
			return g.typeExprInContext(ctx, info.Definition.ReturnType), true
		}
		if binding, ok := ctx.lookup(callee.Name); ok && binding.TypeExpr != nil {
			if fnType, ok := binding.TypeExpr.(*ast.FunctionTypeExpression); ok && fnType != nil && fnType.ReturnType != nil {
				return g.typeExprInContext(ctx, fnType.ReturnType), true
			}
		}
	case *ast.MemberAccessExpression:
		memberIdent, ok := callee.Member.(*ast.Identifier)
		if !ok || memberIdent == nil {
			return nil, false
		}
		if method, ok := g.resolveStaticMethodCall(ctx, callee.Object, memberIdent.Name); ok && method != nil && method.Info != nil && method.Info.Definition != nil && method.Info.Definition.ReturnType != nil {
			return g.typeExprInContext(ctx, method.Info.Definition.ReturnType), true
		}
	}
	return nil, false
}

func (g *generator) inferStaticArrayLiteralTypeExpr(ctx *compileContext, lit *ast.ArrayLiteral) (ast.TypeExpression, bool) {
	if lit == nil {
		return nil, false
	}
	if len(lit.Elements) == 0 {
		return ast.Gen(ast.Ty("Array"), ast.NewWildcardTypeExpression()), true
	}
	var elemExpr ast.TypeExpression
	for _, element := range lit.Elements {
		inferred, ok := g.inferStaticElementTypeExpr(ctx, element)
		if !ok || inferred == nil {
			return ast.Gen(ast.Ty("Array"), ast.NewWildcardTypeExpression()), true
		}
		if elemExpr == nil {
			elemExpr = inferred
			continue
		}
		if typeExpressionToString(elemExpr) != typeExpressionToString(inferred) {
			return ast.Gen(ast.Ty("Array"), ast.NewWildcardTypeExpression()), true
		}
	}
	if elemExpr == nil {
		elemExpr = ast.NewWildcardTypeExpression()
	}
	return ast.Gen(ast.Ty("Array"), elemExpr), true
}

func (g *generator) inferStaticElementTypeExpr(ctx *compileContext, expr ast.Expression) (ast.TypeExpression, bool) {
	if g == nil || expr == nil {
		return nil, false
	}
	switch e := expr.(type) {
	case *ast.Identifier:
		if ctx == nil {
			return nil, false
		}
		if binding, ok := ctx.lookup(e.Name); ok && binding.TypeExpr != nil {
			return g.typeExprInContext(ctx, binding.TypeExpr), true
		}
	case *ast.IntegerLiteral:
		return g.typeExprForGoType(g.inferIntegerLiteralType(e))
	case *ast.FloatLiteral:
		return g.typeExprForGoType(g.inferFloatLiteralType(e))
	case *ast.StringLiteral:
		return ast.Ty("String"), true
	case *ast.BooleanLiteral:
		return ast.Ty("bool"), true
	case *ast.CharLiteral:
		return ast.Ty("char"), true
	case *ast.NilLiteral:
		return ast.Ty("nil"), true
	case *ast.ArrayLiteral:
		return g.inferStaticArrayLiteralTypeExpr(ctx, e)
	}
	return nil, false
}

func (g *generator) staticArrayElementGoTypeForExpr(ctx *compileContext, expr ast.Expression, goType string) string {
	if spec, ok := g.monoArraySpecForGoType(goType); ok && spec != nil {
		return spec.ElemGoType
	}
	typeExpr, ok := g.inferStaticArrayTypeExpr(ctx, expr, goType)
	if !ok || typeExpr == nil {
		return g.staticArrayElemGoType(goType)
	}
	gen, ok := typeExpr.(*ast.GenericTypeExpression)
	if !ok || gen == nil || len(gen.Arguments) != 1 {
		return g.staticArrayElemGoType(goType)
	}
	elemGoType, ok := g.mapTypeExpressionInContext(ctx, gen.Arguments[0])
	if !ok || elemGoType == "" {
		return g.staticArrayElemGoType(goType)
	}
	return elemGoType
}

func (g *generator) staticArrayDefaultNullableResultType(ctx *compileContext, expr ast.Expression, goType string) (string, bool) {
	elemGoType := g.staticArrayElementGoTypeForExpr(ctx, expr, goType)
	if elemGoType == "" {
		return "", false
	}
	if g.isNilableStaticCarrierType(elemGoType) {
		return elemGoType, true
	}
	return g.nativeNullablePointerType(elemGoType)
}

func (g *generator) inferLocalTypeExpr(ctx *compileContext, expr ast.Expression, goType string) (ast.TypeExpression, bool) {
	if typeExpr, ok := g.inferStaticArrayTypeExpr(ctx, expr, goType); ok && typeExpr != nil {
		return typeExpr, true
	}
	if ident, ok := expr.(*ast.Identifier); ok && ident != nil && ctx != nil {
		if binding, found := ctx.lookup(ident.Name); found && binding.TypeExpr != nil {
			return g.typeExprInContext(ctx, binding.TypeExpr), true
		}
	}
	return nil, false
}

func (g *generator) staticArrayResultExprLines(ctx *compileContext, arrayType string, elemExpr string, expected string) ([]string, string, string, bool) {
	if ctx == nil {
		return nil, "", "", false
	}
	if expected == "" {
		expected = "runtime.Value"
	}
	actual := g.staticArrayElemGoType(arrayType)
	if actual == "" {
		return nil, "", "", false
	}
	if actual == "runtime.Value" {
		if expected == "runtime.Value" {
			temp := ctx.newTemp()
			lines := []string{
				fmt.Sprintf("var %s runtime.Value = %s", temp, elemExpr),
				fmt.Sprintf("if %s == nil { %s = runtime.NilValue{} }", temp, temp),
			}
			return lines, temp, "runtime.Value", true
		}
		return g.coerceExpectedStaticExpr(ctx, nil, elemExpr, "runtime.Value", expected)
	}
	if expected == "runtime.Value" {
		runtimeExpr, ok := g.staticArrayElementRuntimeExpr(arrayType, elemExpr)
		if !ok {
			return nil, "", "", false
		}
		return nil, runtimeExpr, "runtime.Value", true
	}
	if expected == actual {
		return nil, elemExpr, actual, true
	}
	return g.coerceExpectedStaticExpr(ctx, nil, elemExpr, actual, expected)
}
