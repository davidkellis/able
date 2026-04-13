package compiler

import (
	"fmt"
	"strings"

	"able/interpreter-go/pkg/ast"
)

func (g *generator) nativeCallableInfoForMethod(method *methodInfo) (*nativeCallableInfo, bool) {
	if g == nil || method == nil || method.Info == nil {
		return nil, false
	}
	paramOffset := 0
	if method.ExpectsSelf {
		paramOffset = 1
	}
	paramExprs := make([]ast.TypeExpression, 0, len(method.Info.Params)-paramOffset)
	paramGoTypes := make([]string, 0, len(method.Info.Params)-paramOffset)
	for idx := paramOffset; idx < len(method.Info.Params); idx++ {
		param := method.Info.Params[idx]
		paramExpr := param.TypeExpr
		if paramExpr == nil {
			fallback, ok := g.typeExprForGoType(param.GoType)
			if !ok {
				return nil, false
			}
			paramExpr = fallback
		}
		paramExprs = append(paramExprs, paramExpr)
		paramGoTypes = append(paramGoTypes, param.GoType)
	}
	var returnExpr ast.TypeExpression
	if method.Info.Definition != nil {
		returnExpr = method.Info.Definition.ReturnType
		if method.TargetType != nil {
			returnExpr = resolveSelfTypeExpr(returnExpr, method.TargetType)
		}
	}
	if returnExpr != nil {
		returnExpr = normalizeTypeExprForPackage(g, method.Info.Package, returnExpr)
	}
	return g.ensureNativeCallableInfoFromSignatureInPackage(method.Info.Package, paramExprs, paramGoTypes, returnExpr, method.Info.ReturnType)
}

func (g *generator) compileNativeBoundMethodValue(ctx *compileContext, objectExpr string, objectType string, method *methodInfo) ([]string, string, string, bool) {
	if g == nil || ctx == nil || objectExpr == "" || objectType == "" || method == nil || method.Info == nil || !method.Info.Compileable {
		return nil, "", "", false
	}
	callableInfo, ok := g.nativeCallableInfoForMethod(method)
	if !ok || callableInfo == nil {
		return nil, "", "", false
	}
	receiverTemp := ctx.newTemp()
	lines := []string{fmt.Sprintf("var %s %s = %s", receiverTemp, objectType, objectExpr)}
	paramParts := make([]string, 0, len(callableInfo.ParamGoTypes))
	argNames := make([]string, 0, len(callableInfo.ParamGoTypes))
	for idx, paramType := range callableInfo.ParamGoTypes {
		name := fmt.Sprintf("arg%d", idx)
		paramParts = append(paramParts, fmt.Sprintf("%s %s", name, paramType))
		argNames = append(argNames, name)
	}
	var callExpr string
	if g.nativeInterfaceInfoForGoType(objectType) != nil {
		callExpr = fmt.Sprintf("%s.%s(%s)", receiverTemp, sanitizeIdent(method.MethodName), strings.Join(argNames, ", "))
	} else {
		args := append([]string{}, argNames...)
		if method.ExpectsSelf {
			args = append([]string{receiverTemp}, args...)
		}
		callExpr = fmt.Sprintf("__able_compiled_%s(%s)", method.Info.GoName, strings.Join(args, ", "))
	}
	bodyParts := append([]string{}, g.inlineRuntimeEnvSwapLinesForPackage(ctx.packageName)...)
	bodyParts = append(bodyParts, fmt.Sprintf("return %s", callExpr))
	callableExpr := fmt.Sprintf("%s(func(%s) (%s, *__ableControl) { %s })", callableInfo.GoType, strings.Join(paramParts, ", "), callableInfo.ReturnGoType, strings.Join(bodyParts, "; "))
	return lines, callableExpr, callableInfo.GoType, true
}

func (g *generator) compileNativeInterfaceBoundMethodValue(ctx *compileContext, objectExpr string, objectType string, method *nativeInterfaceMethod) ([]string, string, string, bool) {
	if g == nil || ctx == nil || objectExpr == "" || objectType == "" || method == nil {
		return nil, "", "", false
	}
	callableInfo, ok := g.ensureNativeCallableInfoFromSignatureInPackage(ctx.packageName, method.ParamTypeExprs, method.ParamGoTypes, method.ReturnTypeExpr, method.ReturnGoType)
	if !ok || callableInfo == nil {
		return nil, "", "", false
	}
	receiverTemp := ctx.newTemp()
	lines := []string{fmt.Sprintf("var %s %s = %s", receiverTemp, objectType, objectExpr)}
	paramParts := make([]string, 0, len(callableInfo.ParamGoTypes))
	argNames := make([]string, 0, len(callableInfo.ParamGoTypes))
	for idx, paramType := range callableInfo.ParamGoTypes {
		name := fmt.Sprintf("arg%d", idx)
		paramParts = append(paramParts, fmt.Sprintf("%s %s", name, paramType))
		argNames = append(argNames, name)
	}
	callExpr := fmt.Sprintf("%s.%s(%s)", receiverTemp, method.GoName, strings.Join(argNames, ", "))
	bodyParts := append([]string{}, g.inlineRuntimeEnvSwapLinesForPackage(ctx.packageName)...)
	bodyParts = append(bodyParts, fmt.Sprintf("return %s", callExpr))
	callableExpr := fmt.Sprintf("%s(func(%s) (%s, *__ableControl) { %s })", callableInfo.GoType, strings.Join(paramParts, ", "), callableInfo.ReturnGoType, strings.Join(bodyParts, "; "))
	return lines, callableExpr, callableInfo.GoType, true
}
