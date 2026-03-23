package compiler

import (
	"fmt"
	"strings"

	"able/interpreter-go/pkg/ast"
)

func (g *generator) compileStaticNamedFunctionCall(ctx *compileContext, call *ast.FunctionCall, expected string, calleeName string, info *functionInfo, overload *overloadInfo, callNode string) ([]string, string, string, bool) {
	if g == nil || ctx == nil || call == nil {
		return nil, "", "", false
	}
	if info != nil && info.Compileable {
		info = g.concreteFunctionCallInfo(ctx, call, info, expected)
		needsRuntimeValue := expected == "runtime.Value" && info.ReturnType != "runtime.Value"
		needsExpect := expected != "" && expected != "runtime.Value" && info.ReturnType == "runtime.Value"
		needsAnyConv := expected != "" && expected != "any" && info.ReturnType == "any"
		needsStaticCoerce := expected != "" && expected != "runtime.Value" && expected != "any" && g.canCoerceStaticExpr(expected, info.ReturnType)
		if !g.typeMatches(expected, info.ReturnType) && !needsRuntimeValue && !needsExpect && !needsAnyConv && !needsStaticCoerce {
			ctx.setReason("call return type mismatch")
			return nil, "", "", false
		}
		optionalLast := g.hasOptionalLastParam(info)
		if len(call.Arguments) != len(info.Params) {
			if !(optionalLast && len(call.Arguments) == len(info.Params)-1) {
				ctx.setReason("call arity mismatch")
				return nil, "", "", false
			}
		}
		missingOptional := optionalLast && len(call.Arguments) == len(info.Params)-1
		if missingOptional && len(info.Params) > 0 {
			lastType := info.Params[len(info.Params)-1].GoType
			if lastType != "runtime.Value" && lastType != "any" {
				ctx.setReason("call arity mismatch")
				return nil, "", "", false
			}
		}
		args := make([]string, 0, len(call.Arguments))
		preLines := make([]string, 0, len(call.Arguments))
		postLines := make([]string, 0, len(call.Arguments))
		var writebackIdents []string
		for idx, arg := range call.Arguments {
			param := info.Params[idx]
			paramGoType := param.GoType
			expectedArgType := g.staticParamCarrierType(ctx, param)
			compileExpectedArgType := expectedArgType
			if g.nativeUnionInfoForGoType(expectedArgType) != nil {
				compileExpectedArgType = ""
			}
			if ifaceType, ok := g.interfaceTypeExpr(param.TypeExpr); ok && (expectedArgType == "" || expectedArgType == "runtime.Value" || expectedArgType == "any") {
				if ifaceInfo, ok := g.ensureNativeInterfaceInfo(ctx.packageName, ifaceType); ok && ifaceInfo != nil && ifaceInfo.GoType != "" {
					expectedArgType = ifaceInfo.GoType
				}
			}
			if g.typeCategory(paramGoType) == "struct" {
				if ident, ok := arg.(*ast.Identifier); ok && ident != nil {
					if binding, ok := ctx.lookup(ident.Name); ok && binding.GoType == "runtime.Value" {
						runtimeTemp := ctx.newTemp()
						preLines = append(preLines, fmt.Sprintf("%s := %s", runtimeTemp, binding.GoName))
						convLines, structExpr, ok := g.lowerExpectRuntimeValue(ctx, runtimeTemp, paramGoType)
						if !ok {
							ctx.setReason("call argument unsupported")
							return nil, "", "", false
						}
						preLines = append(preLines, convLines...)
						structTemp := ctx.newTemp()
						preLines = append(preLines, fmt.Sprintf("%s := %s", structTemp, structExpr))
						args = append(args, structTemp)
						baseName, ok := g.structBaseName(paramGoType)
						if !ok {
							baseName = strings.TrimPrefix(paramGoType, "*")
						}
						transferLines, ok := g.lowerControlTransfer(ctx, g.runtimeErrorControlExpr(callNode, "err"))
						if !ok {
							return nil, "", "", false
						}
						postLines = append(postLines, fmt.Sprintf("if err := __able_struct_%s_apply(__able_runtime, %s, %s); err != nil {", baseName, runtimeTemp, structTemp))
						postLines = append(postLines, indentLines(transferLines, 1)...)
						postLines = append(postLines, "}")
						writebackIdents = append(writebackIdents, ident.Name)
						continue
					}
				}
			}
			argLines, expr, exprType, ok := g.compileExprLinesWithExpectedTypeExpr(ctx, arg, compileExpectedArgType, param.TypeExpr)
			if !ok {
				return nil, "", "", false
			}
			preLines = append(preLines, argLines...)
			argExpr := expr
			argType := exprType
			if paramGoType == "runtime.Value" && argType != "runtime.Value" {
				argConvLines, valueExpr, ok := g.lowerRuntimeValue(ctx, argExpr, argType)
				if !ok {
					ctx.setReason("call argument unsupported")
					return nil, "", "", false
				}
				preLines = append(preLines, argConvLines...)
				argExpr = valueExpr
				argType = "runtime.Value"
			}
			if ifaceType, ok := g.interfaceTypeExpr(param.TypeExpr); ok && paramGoType == "runtime.Value" {
				ifaceLines, coerced, ok := g.interfaceArgExprLines(ctx, argExpr, ifaceType, calleeName, ctx.genericNames)
				if !ok {
					ctx.setReason("interface argument unsupported")
					return nil, "", "", false
				}
				preLines = append(preLines, ifaceLines...)
				argExpr = coerced
			} else if paramGoType != "" && paramGoType != "any" && argType != paramGoType {
				coerceLines, coercedExpr, coercedType, ok := g.prepareStaticCallArg(ctx, argExpr, argType, paramGoType)
				if !ok {
					ctx.setReason("call argument type mismatch")
					return nil, "", "", false
				}
				preLines = append(preLines, coerceLines...)
				argExpr = coercedExpr
				argType = coercedType
			}
			args = append(args, argExpr)
		}
		if missingOptional {
			lastType := info.Params[len(info.Params)-1].GoType
			if lastType == "any" {
				args = append(args, "nil")
			} else {
				args = append(args, "runtime.NilValue{}")
			}
		}
		callExpr := fmt.Sprintf("__able_compiled_%s(%s)", info.GoName, strings.Join(args, ", "))
		resultTemp := ctx.newTemp()
		controlTemp := ctx.newTemp()
		lines := make([]string, 0, len(preLines)+len(postLines)+4)
		lines = append(lines, fmt.Sprintf("__able_push_call_frame(%s)", callNode))
		lines = append(lines, preLines...)
		lines = append(lines, fmt.Sprintf("%s, %s := %s", resultTemp, controlTemp, callExpr))
		lines = append(lines, "__able_pop_call_frame()")
		controlLines, ok := g.lowerControlCheck(ctx, controlTemp)
		if !ok {
			return nil, "", "", false
		}
		lines = append(lines, controlLines...)
		lines = append(lines, postLines...)
		if len(writebackIdents) > 0 && ctx.originExtractions != nil {
			for _, name := range writebackIdents {
				delete(ctx.originExtractions, name)
			}
		}
		if needsRuntimeValue {
			convLines, converted, ok := g.lowerRuntimeValue(ctx, resultTemp, info.ReturnType)
			if !ok {
				ctx.setReason("call return type mismatch")
				return nil, "", "", false
			}
			lines = append(lines, convLines...)
			return lines, converted, "runtime.Value", true
		}
		if needsExpect {
			convLines, converted, ok := g.lowerExpectRuntimeValue(ctx, resultTemp, expected)
			if !ok {
				ctx.setReason("call return type mismatch")
				return nil, "", "", false
			}
			lines = append(lines, convLines...)
			return lines, converted, expected, true
		}
		if needsAnyConv {
			if expected == "runtime.Value" {
				convTemp := ctx.newTemp()
				lines = append(lines, fmt.Sprintf("%s := __able_any_to_value(%s)", convTemp, resultTemp))
				return lines, convTemp, "runtime.Value", true
			}
			anyTemp := ctx.newTemp()
			lines = append(lines, fmt.Sprintf("%s := __able_any_to_value(%s)", anyTemp, resultTemp))
			convLines, converted, ok := g.lowerExpectRuntimeValue(ctx, anyTemp, expected)
			if !ok {
				ctx.setReason("call return type mismatch")
				return nil, "", "", false
			}
			lines = append(lines, convLines...)
			return lines, converted, expected, true
		}
		if needsStaticCoerce {
			return g.lowerCoerceExpectedStaticExpr(ctx, lines, resultTemp, info.ReturnType, expected)
		}
		return lines, resultTemp, info.ReturnType, true
	}
	if overload != nil {
		return g.compileResolvedOverloadCall(ctx, call, expected, overload.Package, overload.Name, callNode)
	}
	return nil, "", "", false
}
