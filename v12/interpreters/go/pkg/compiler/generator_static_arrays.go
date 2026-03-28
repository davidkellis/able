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
		capacityExpr = g.staticSliceLenExpr(valuesExpr)
	}
	valuesTemp := ctx.newTemp()
	arrayTemp := ctx.newTemp()
	if spec, ok := g.monoArraySpecForGoType(arrayType); ok && spec != nil {
		lines := []string{
			fmt.Sprintf("%s := make([]%s, %s, %s)", valuesTemp, spec.ElemGoType, g.staticSliceLenExpr(valuesExpr), capacityExpr),
			fmt.Sprintf("copy(%s, %s)", valuesTemp, valuesExpr),
			fmt.Sprintf("%s := &%s{Elements: %s}", arrayTemp, spec.GoName, valuesTemp),
			fmt.Sprintf("%s(%s)", spec.SyncHelper, arrayTemp),
		}
		return lines, arrayTemp, true
	}
	if g.isArrayStructType(arrayType) {
		lines := []string{
			fmt.Sprintf("%s := make([]runtime.Value, %s, %s)", valuesTemp, g.staticSliceLenExpr(valuesExpr), capacityExpr),
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
			return g.lowerNormalizedTypeExpr(ctx, binding.TypeExpr), true
		}
	case *ast.MemberAccessExpression:
		if inferred, ok := g.inferMemberAccessTypeExpr(ctx, e); ok && inferred != nil {
			return g.lowerNormalizedTypeExpr(ctx, inferred), true
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
		if helperExpr := g.runtimeHelperResultTypeExpr(callee.Name); helperExpr != nil {
			return g.lowerNormalizedTypeExpr(ctx, helperExpr), true
		}
		if info, _, ok := g.resolveStaticCallable(ctx, callee.Name); ok && info != nil && info.Definition != nil && info.Definition.ReturnType != nil {
			info = g.concreteFunctionCallInfo(ctx, call, info, "")
			if returnExpr := g.functionReturnTypeExpr(info); returnExpr != nil {
				return returnExpr, true
			}
		}
		if binding, ok := ctx.lookup(callee.Name); ok && binding.TypeExpr != nil {
			if fnType, ok := binding.TypeExpr.(*ast.FunctionTypeExpression); ok && fnType != nil && fnType.ReturnType != nil {
				return g.lowerNormalizedTypeExpr(ctx, fnType.ReturnType), true
			}
		}
	case *ast.MemberAccessExpression:
		memberIdent, ok := callee.Member.(*ast.Identifier)
		if !ok || memberIdent == nil {
			return nil, false
		}
		if typeIdent, ok := callee.Object.(*ast.Identifier); ok && typeIdent != nil && typeIdent.Name == "Array" {
			switch memberIdent.Name {
			case "new", "with_capacity":
				return g.staticArrayFactoryResultTypeExpr(ctx), true
			}
		}
		if receiverTypeExpr, ok := g.inferExpressionTypeExpr(ctx, callee.Object, ""); ok && receiverTypeExpr != nil {
			if resultExpr, ok := g.staticArrayMethodResultTypeExpr(ctx, receiverTypeExpr, memberIdent.Name); ok && resultExpr != nil {
				return g.wrapSafeNavigationTypeExpr(ctx, receiverTypeExpr, callee.Safe, resultExpr), true
			}
			if receiverGoType, ok := g.lowerCarrierType(ctx, receiverTypeExpr); ok && receiverGoType != "" {
				if method, ok := g.nativeInterfaceGenericMethodForGoType(receiverGoType, memberIdent.Name); ok && method != nil {
					if _, _, returnExpr, _, _, ok := g.inferNativeInterfaceGenericMethodShape(ctx, call, receiverGoType, method, ""); ok && returnExpr != nil {
						return g.wrapSafeNavigationTypeExpr(ctx, receiverTypeExpr, callee.Safe, returnExpr), true
					}
				}
				if method, ok := g.nativeInterfaceMethodForGoType(receiverGoType, memberIdent.Name); ok && method != nil && method.ReturnTypeExpr != nil {
					return g.wrapSafeNavigationTypeExpr(ctx, receiverTypeExpr, callee.Safe, g.lowerNormalizedTypeExpr(ctx, method.ReturnTypeExpr)), true
				}
				if method := g.methodForReceiver(receiverGoType, memberIdent.Name); method != nil && method.Info != nil {
					method = g.concreteMethodCallInfo(ctx, call, method, callee.Object, receiverGoType, "")
					if returnExpr := g.functionReturnTypeExpr(method.Info); returnExpr != nil {
						return g.wrapSafeNavigationTypeExpr(ctx, receiverTypeExpr, callee.Safe, returnExpr), true
					}
				}
				if method := g.compileableInterfaceMethodForConcreteReceiver(receiverGoType, memberIdent.Name); method != nil && method.Info != nil {
					method = g.concreteMethodCallInfo(ctx, call, method, callee.Object, receiverGoType, "")
					if returnExpr := g.functionReturnTypeExpr(method.Info); returnExpr != nil {
						return g.wrapSafeNavigationTypeExpr(ctx, receiverTypeExpr, callee.Safe, returnExpr), true
					}
				}
				if _, method, ok := g.nativeInterfaceGenericMethodForConcreteReceiver(receiverGoType, receiverTypeExpr, memberIdent.Name); ok && method != nil {
					if _, _, returnExpr, _, _, ok := g.inferNativeInterfaceGenericMethodShape(ctx, call, receiverGoType, method, ""); ok && returnExpr != nil {
						return g.wrapSafeNavigationTypeExpr(ctx, receiverTypeExpr, callee.Safe, returnExpr), true
					}
				}
			}
		}
		if method, ok := g.lowerResolveStaticMethod(ctx, callee.Object, memberIdent.Name); ok && method != nil && method.Info != nil {
			method = g.concreteStaticMethodCallInfo(ctx, call, method, callee.Object, "")
			if returnExpr := g.functionReturnTypeExpr(method.Info); returnExpr != nil {
				if receiverTypeExpr, ok := g.inferExpressionTypeExpr(ctx, callee.Object, ""); ok && receiverTypeExpr != nil {
					return g.wrapSafeNavigationTypeExpr(ctx, receiverTypeExpr, callee.Safe, returnExpr), true
				}
				return returnExpr, true
			}
		}
	}
	return nil, false
}

func (g *generator) staticArrayMethodResultTypeExpr(ctx *compileContext, receiverTypeExpr ast.TypeExpression, methodName string) (ast.TypeExpression, bool) {
	if g == nil || receiverTypeExpr == nil {
		return nil, false
	}
	receiverTypeExpr = g.lowerNormalizedTypeExpr(ctx, receiverTypeExpr)
	baseName, ok := typeExprBaseName(receiverTypeExpr)
	if !ok || baseName != "Array" {
		return nil, false
	}
	elemTypeExpr := ast.TypeExpression(ast.NewWildcardTypeExpression())
	if generic, ok := receiverTypeExpr.(*ast.GenericTypeExpression); ok && generic != nil && len(generic.Arguments) == 1 && generic.Arguments[0] != nil {
		elemTypeExpr = g.lowerNormalizedTypeExpr(ctx, generic.Arguments[0])
	}
	switch methodName {
	case "len", "size", "capacity":
		return ast.Ty("i32"), true
	case "is_empty":
		return ast.Ty("bool"), true
	case "push", "push_all", "clear", "write_slot", "reserve", "refresh_metadata":
		return ast.Ty("void"), true
	case "read_slot":
		return elemTypeExpr, true
	case "get", "pop", "first", "last":
		return ast.NewNullableTypeExpression(elemTypeExpr), true
	case "set":
		return ast.NewResultTypeExpression(ast.Ty("nil")), true
	case "clone_shallow":
		return receiverTypeExpr, true
	default:
		return nil, false
	}
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
			return g.lowerNormalizedTypeExpr(ctx, binding.TypeExpr), true
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
	elemGoType, ok := g.lowerCarrierType(ctx, gen.Arguments[0])
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

func (g *generator) staticArrayPropagationResultType(ctx *compileContext, expr ast.Expression, goType string) (string, bool) {
	elemGoType := g.staticArrayElementGoTypeForExpr(ctx, expr, goType)
	if elemGoType == "" {
		return "", false
	}
	if innerType, ok := g.nativeNullableValueInnerType(elemGoType); ok {
		return innerType, true
	}
	return elemGoType, true
}

func (g *generator) inferLocalTypeExpr(ctx *compileContext, expr ast.Expression, goType string) (ast.TypeExpression, bool) {
	if typeExpr, ok := g.inferStaticArrayTypeExpr(ctx, expr, goType); ok && typeExpr != nil {
		return typeExpr, true
	}
	if typeExpr, ok := g.inferStaticElementTypeExpr(ctx, expr); ok && typeExpr != nil {
		return typeExpr, true
	}
	if lit, ok := expr.(*ast.StructLiteral); ok && lit != nil {
		if inferred := g.staticStructLiteralTypeExpr(ctx, lit, ""); inferred != nil {
			return inferred, true
		}
	}
	if call, ok := expr.(*ast.FunctionCall); ok && call != nil {
		if inferred, ok := g.inferStaticCallResultTypeExpr(ctx, call); ok && inferred != nil {
			return inferred, true
		}
	}
	if ident, ok := expr.(*ast.Identifier); ok && ident != nil && ctx != nil {
		if binding, found := ctx.lookup(ident.Name); found {
			if inferred := g.inferBindingTypeExpr(ctx, binding); inferred != nil {
				return inferred, true
			}
		}
	}
	if member, ok := expr.(*ast.MemberAccessExpression); ok && member != nil {
		if inferred, ok := g.inferMemberAccessTypeExpr(ctx, member); ok && inferred != nil {
			return inferred, true
		}
	}
	if inferred := g.inferredExpressionTypeExpr(ctx, expr); inferred != nil {
		inferred = g.lowerNormalizedTypeExpr(ctx, inferred)
		if !g.typeExprFullyBound(ctx.packageName, inferred) {
			if goType == "" || goType == "runtime.Value" || goType == "any" {
				return nil, false
			}
		} else {
			return inferred, true
		}
	}
	if goType != "" {
		if inferred, ok := g.typeExprForGoType(goType); ok && inferred != nil {
			return g.lowerNormalizedTypeExpr(ctx, inferred), true
		}
	}
	return nil, false
}

func (g *generator) inferBindingTypeExpr(ctx *compileContext, binding paramInfo) ast.TypeExpression {
	if g == nil || ctx == nil {
		return nil
	}
	var boundExpr ast.TypeExpression
	if binding.TypeExpr != nil {
		boundExpr = g.lowerNormalizedTypeExpr(ctx, ctx.substituteTypeBindings(binding.TypeExpr))
	}
	if binding.GoType == "" || binding.GoType == "runtime.Value" || binding.GoType == "any" {
		return boundExpr
	}
	carrierExpr, ok := g.typeExprForGoType(binding.GoType)
	if !ok || carrierExpr == nil {
		return boundExpr
	}
	carrierExpr = g.lowerNormalizedTypeExpr(ctx, carrierExpr)
	if boundExpr == nil {
		return carrierExpr
	}
	if typeExpressionToString(boundExpr) == typeExpressionToString(carrierExpr) {
		return boundExpr
	}
	boundBase, boundOK := typeExprBaseName(boundExpr)
	carrierBase, carrierOK := typeExprBaseName(carrierExpr)
	if boundOK && carrierOK && boundBase == carrierBase {
		if !g.typeExprHasWildcard(boundExpr) && g.typeExprHasWildcard(carrierExpr) {
			return boundExpr
		}
		if g.typeExprFullyBound(ctx.packageName, carrierExpr) && !g.typeExprFullyBound(ctx.packageName, boundExpr) {
			return carrierExpr
		}
		if g.typeExprFullyBound(ctx.packageName, boundExpr) && !g.typeExprFullyBound(ctx.packageName, carrierExpr) {
			return boundExpr
		}
		return boundExpr
	}
	return boundExpr
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
		return g.lowerCoerceExpectedStaticExpr(ctx, nil, elemExpr, "runtime.Value", expected)
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
	return g.lowerCoerceExpectedStaticExpr(ctx, nil, elemExpr, actual, expected)
}
