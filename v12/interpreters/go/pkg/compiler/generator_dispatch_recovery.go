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

func (g *generator) dispatchReceiverTypeExpr(ctx *compileContext, expr ast.Expression) ast.TypeExpression {
	if g == nil || ctx == nil || expr == nil {
		return nil
	}
	if ident, ok := expr.(*ast.Identifier); ok && ident != nil && ident.Name != "" {
		if binding, ok := ctx.lookup(ident.Name); ok {
			if inferred := g.inferBindingTypeExpr(ctx, binding); inferred != nil {
				return g.lowerNormalizedTypeExpr(ctx, inferred)
			}
			if binding.GoType != "" && binding.GoType != "runtime.Value" && binding.GoType != "any" && !g.isVoidType(binding.GoType) {
				if inferred, ok := g.typeExprForGoType(binding.GoType); ok && inferred != nil {
					return g.lowerNormalizedTypeExpr(ctx, inferred)
				}
			}
		}
	}
	inferred, ok := g.inferExpressionTypeExpr(ctx, expr, "")
	if !ok || inferred == nil {
		return nil
	}
	return g.lowerNormalizedTypeExpr(ctx, inferred)
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

func (g *generator) inferredDispatchReceiverType(ctx *compileContext, expr ast.Expression) string {
	if g == nil || ctx == nil || expr == nil {
		return ""
	}
	inferred := g.dispatchReceiverTypeExpr(ctx, expr)
	if inferred == nil {
		return ""
	}
	if recovered, ok := g.joinCarrierTypeFromTypeExpr(ctx, inferred); ok && recovered != "" && recovered != "runtime.Value" && recovered != "any" && !g.isVoidType(recovered) {
		return recovered
	}
	if recovered, ok := g.lowerCarrierType(ctx, inferred); ok && recovered != "" && recovered != "runtime.Value" && recovered != "any" && !g.isVoidType(recovered) {
		return recovered
	}
	return ""
}

func (g *generator) compileDispatchReceiverExpr(ctx *compileContext, expr ast.Expression) ([]string, string, string, bool) {
	return g.compileDispatchReceiverExprWithExpectedTypeExpr(ctx, expr, "", nil)
}

func (g *generator) compileDispatchReceiverExprWithExpectedTypeExpr(ctx *compileContext, expr ast.Expression, expectedGoType string, expectedTypeExpr ast.TypeExpression) ([]string, string, string, bool) {
	if g == nil || ctx == nil || expr == nil {
		return nil, "", "", false
	}
	expectedType := expectedGoType
	if expectedType == "" && expectedTypeExpr != nil {
		if lowered, ok := g.lowerCarrierType(ctx, expectedTypeExpr); ok && lowered != "" {
			expectedType = lowered
		}
	}
	if expectedType == "" {
		expectedType = g.inferredDispatchReceiverType(ctx, expr)
	}
	var guidedLines []string
	var guidedExpr string
	var guidedType string
	guidedOK := false
	if expectedType != "" || expectedTypeExpr != nil {
		guidedLines, guidedExpr, guidedType, guidedOK = g.compileExprLinesWithExpectedTypeExpr(ctx, expr, expectedType, expectedTypeExpr)
		if guidedOK {
			if recoverLines, recoveredExpr, recoveredType, recovered := g.recoverDispatchExpr(ctx, expr, guidedExpr, guidedType); recovered {
				guidedLines = append(guidedLines, recoverLines...)
				guidedExpr = recoveredExpr
				guidedType = recoveredType
			}
			if guidedType != "" && guidedType != "runtime.Value" && guidedType != "any" && !g.isVoidType(guidedType) {
				return guidedLines, guidedExpr, guidedType, true
			}
		}
	}
	lines, compiledExpr, compiledType, ok := g.compileExprLines(ctx, expr, "")
	if ok {
		if recoverLines, recoveredExpr, recoveredType, recovered := g.recoverDispatchExpr(ctx, expr, compiledExpr, compiledType); recovered {
			lines = append(lines, recoverLines...)
			compiledExpr = recoveredExpr
			compiledType = recoveredType
		}
		if compiledType != "" && compiledType != "runtime.Value" && compiledType != "any" && !g.isVoidType(compiledType) {
			return lines, compiledExpr, compiledType, true
		}
	}
	if ok {
		return lines, compiledExpr, compiledType, true
	}
	if guidedOK {
		return guidedLines, guidedExpr, guidedType, true
	}
	if expectedType == "" && expectedTypeExpr == nil {
		return nil, "", "", false
	}
	return nil, "", "", false
}

func (g *generator) preferredDispatchReceiverTypeExpr(ctx *compileContext, call *ast.FunctionCall, expr ast.Expression, methodName string, expected string) ast.TypeExpression {
	if g == nil || ctx == nil || call == nil || expr == nil || methodName == "" {
		return nil
	}
	if ident, ok := expr.(*ast.Identifier); ok && ident != nil && ident.Name != "" {
		if binding, ok := ctx.lookup(ident.Name); ok && binding.GoType != "" && binding.GoType != "runtime.Value" && binding.GoType != "any" {
			if ifaceInfo := g.nativeInterfaceInfoForGoType(binding.GoType); ifaceInfo != nil {
				if _, ok := g.nativeInterfaceMethodForGoType(ifaceInfo.GoType, methodName); ok {
					if carrierExpr, ok := g.typeExprForGoType(binding.GoType); ok && carrierExpr != nil {
						return g.lowerNormalizedTypeExpr(ctx, carrierExpr)
					}
				}
				if _, ok := g.nativeInterfaceGenericMethodForGoType(ifaceInfo.GoType, methodName); ok {
					if carrierExpr, ok := g.typeExprForGoType(binding.GoType); ok && carrierExpr != nil {
						return g.lowerNormalizedTypeExpr(ctx, carrierExpr)
					}
				}
			}
			if binding.GoType == "runtime.ErrorValue" {
				switch methodName {
				case "message", "cause":
					if carrierExpr, ok := g.typeExprForGoType(binding.GoType); ok && carrierExpr != nil {
						return g.lowerNormalizedTypeExpr(ctx, carrierExpr)
					}
				}
			}
		}
	}
	receiverTypeExpr := g.dispatchReceiverTypeExpr(ctx, expr)
	if receiverTypeExpr == nil {
		return nil
	}
	receiverTypeExpr = normalizeTypeExprForPackage(g, ctx.packageName, receiverTypeExpr)
	if g.typeExprFullyBound(ctx.packageName, receiverTypeExpr) {
		if ifaceInfo, ok := g.ensureNativeInterfaceInfo(ctx.packageName, receiverTypeExpr); ok && ifaceInfo != nil && ifaceInfo.GoType != "" {
			if _, ok := g.nativeInterfaceMethodForGoType(ifaceInfo.GoType, methodName); ok {
				return receiverTypeExpr
			}
			if _, ok := g.nativeInterfaceGenericMethodForGoType(ifaceInfo.GoType, methodName); ok {
				return receiverTypeExpr
			}
		}
	}
	receiverGoType, ok := g.lowerCarrierType(ctx, receiverTypeExpr)
	if !ok || receiverGoType == "" || receiverGoType == "runtime.Value" || receiverGoType == "any" {
		return nil
	}
	method := g.methodForReceiver(receiverGoType, methodName)
	if method == nil {
		return nil
	}
	if concreteReceiverExpr, ok := g.staticReceiverTypeExpr(ctx, expr, receiverGoType); ok && concreteReceiverExpr != nil {
		concreteReceiverExpr = normalizeTypeExprForPackage(g, ctx.packageName, concreteReceiverExpr)
		if g.typeExprFullyBound(ctx.packageName, concreteReceiverExpr) {
			if concreteGoType, ok := g.lowerCarrierType(ctx, concreteReceiverExpr); ok &&
				concreteGoType != "" &&
				concreteGoType != receiverGoType &&
				concreteGoType != "runtime.Value" &&
				concreteGoType != "any" {
				if concreteMethod := g.methodForReceiver(concreteGoType, methodName); concreteMethod != nil {
					currentScore := g.receiverMethodMatchScore(method.ReceiverType, receiverGoType)
					concreteScore := g.receiverMethodMatchScore(concreteMethod.ReceiverType, concreteGoType)
					if concreteScore > currentScore {
						return concreteReceiverExpr
					}
				}
			}
		}
	}
	targetTypeExpr := g.preferredNominalMethodTargetTypeExpr(ctx, call, method, receiverTypeExpr, expected)
	if targetTypeExpr == nil {
		return nil
	}
	targetTypeExpr = normalizeTypeExprForPackage(g, ctx.packageName, targetTypeExpr)
	if !g.typeExprFullyBound(ctx.packageName, targetTypeExpr) {
		return nil
	}
	targetGoType, ok := g.lowerCarrierType(ctx, targetTypeExpr)
	if !ok || targetGoType == "" || targetGoType == receiverGoType || targetGoType == "runtime.Value" || targetGoType == "any" {
		return nil
	}
	targetMethod := g.methodForReceiver(targetGoType, methodName)
	if targetMethod == nil {
		return nil
	}
	currentScore := g.receiverMethodMatchScore(method.ReceiverType, receiverGoType)
	targetScore := g.receiverMethodMatchScore(targetMethod.ReceiverType, targetGoType)
	if targetScore < currentScore {
		return nil
	}
	if targetMethod.Info == method.Info && targetScore <= currentScore {
		return nil
	}
	if targetScore == currentScore &&
		targetMethod.Info != nil &&
		method.Info != nil &&
		targetMethod.Info.GoName == method.Info.GoName {
		return nil
	}
	return targetTypeExpr
}
