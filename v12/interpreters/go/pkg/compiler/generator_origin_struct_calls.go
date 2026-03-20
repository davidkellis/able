package compiler

import (
	"fmt"

	"able/interpreter-go/pkg/ast"
)

func safeNilReturnExpr(expected string) string {
	if expected == "struct{}" {
		return "struct{}{}"
	}
	if expected == "any" {
		return "nil"
	}
	return "runtime.NilValue{}"
}

// compileOriginStructMethodCall resolves a method call on a runtime.Value variable
// whose underlying struct type is known (via OriginGoType). It extracts the struct,
// calls the compiled method directly, and writes back if the object is addressable.
func (g *generator) compileOriginStructMethodCall(ctx *compileContext, call *ast.FunctionCall, expected string, callee *ast.MemberAccessExpression, objExpr string, methodName string, callNode string) ([]string, string, string, bool) {
	if callee == nil || callee.Object == nil {
		return nil, "", "", false
	}
	objIdent, ok := callee.Object.(*ast.Identifier)
	if !ok || objIdent == nil || objIdent.Name == "" {
		return nil, "", "", false
	}
	info, ok := ctx.lookup(objIdent.Name)
	if !ok || info.OriginGoType == "" {
		return nil, "", "", false
	}
	originType := info.OriginGoType
	method := g.methodForReceiver(originType, methodName)
	if method == nil {
		return nil, "", "", false
	}
	baseName, ok := g.structBaseName(originType)
	if !ok {
		return nil, "", "", false
	}
	var lines []string
	var extractTemp string
	cacheHit := false
	if cached, ok := ctx.originExtractions[objIdent.Name]; ok {
		extractTemp = cached
		cacheHit = true
	} else {
		extractTemp = ctx.newTemp()
		extractErrTemp := ctx.newTemp()
		lines = []string{
			fmt.Sprintf("%s, %s := __able_struct_%s_from(%s)", extractTemp, extractErrTemp, baseName, objExpr),
		}
		controlTemp := ctx.newTemp()
		lines = append(lines, fmt.Sprintf("%s := __able_control_from_error(%s)", controlTemp, extractErrTemp))
		controlLines, ok := g.controlCheckLines(ctx, controlTemp)
		if !ok {
			return nil, "", "", false
		}
		lines = append(lines, controlLines...)
	}
	if intrLines, expr, retType, ok := g.compileArrayMethodIntrinsicCall(ctx, nil, extractTemp, originType, methodName, call.Arguments, expected, callNode); ok {
		if !cacheHit {
			if ctx.originExtractions == nil {
				ctx.originExtractions = make(map[string]string)
			}
			ctx.originExtractions[objIdent.Name] = extractTemp
		}
		lines = append(lines, intrLines...)
		if g.isAddressableMemberObject(callee.Object) {
			wbTemp := ctx.newTemp()
			wbErrTemp := ctx.newTemp()
			lines = append(lines,
				fmt.Sprintf("%s, %s := __able_struct_%s_to(__able_runtime, %s)", wbTemp, wbErrTemp, baseName, extractTemp),
			)
			controlTemp := ctx.newTemp()
			lines = append(lines, fmt.Sprintf("%s := __able_control_from_error(%s)", controlTemp, wbErrTemp))
			controlLines, ok := g.controlCheckLines(ctx, controlTemp)
			if !ok {
				return nil, "", "", false
			}
			lines = append(lines, controlLines...)
			lines = append(lines, fmt.Sprintf("%s = %s", objExpr, wbTemp))
			delete(ctx.originExtractions, objIdent.Name)
		}
		return lines, expr, retType, true
	}
	methodLines, resultExpr, resultType, ok := g.compileResolvedMethodCall(ctx, call, expected, method, extractTemp, originType, callNode)
	if !ok {
		return nil, "", "", false
	}
	if !cacheHit {
		if ctx.originExtractions == nil {
			ctx.originExtractions = make(map[string]string)
		}
		ctx.originExtractions[objIdent.Name] = extractTemp
	}
	lines = append(lines, methodLines...)
	if g.isAddressableMemberObject(callee.Object) {
		wbTemp := ctx.newTemp()
		wbErrTemp := ctx.newTemp()
		lines = append(lines,
			fmt.Sprintf("%s, %s := __able_struct_%s_to(__able_runtime, %s)", wbTemp, wbErrTemp, baseName, extractTemp),
		)
		controlTemp := ctx.newTemp()
		lines = append(lines, fmt.Sprintf("%s := __able_control_from_error(%s)", controlTemp, wbErrTemp))
		controlLines, ok := g.controlCheckLines(ctx, controlTemp)
		if !ok {
			return nil, "", "", false
		}
		lines = append(lines, controlLines...)
		lines = append(lines, fmt.Sprintf("%s = %s", objExpr, wbTemp))
		delete(ctx.originExtractions, objIdent.Name)
	}
	return lines, resultExpr, resultType, true
}
