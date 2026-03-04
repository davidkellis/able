package compiler

import (
	"fmt"
	"strings"

	"able/interpreter-go/pkg/ast"
)

func (g *generator) compileFunctionCall(ctx *compileContext, call *ast.FunctionCall, expected string) (string, string, bool) {
	if call == nil {
		ctx.setReason("missing function call")
		return "", "", false
	}
	callNode := g.diagNodeName(call, "*ast.FunctionCall", "call")
	if len(call.TypeArguments) > 0 {
		if callee, ok := call.Callee.(*ast.Identifier); ok && callee != nil {
			// Generic calls on local values (e.g. generic lambdas) must call the
			// bound value, not global name lookup.
			if _, ok := ctx.lookup(callee.Name); ok {
				return g.compileDynamicCall(ctx, call, expected, "", callNode)
			}
			if !g.hasDynamicFeature && !g.mayResolveStaticNamedCall(ctx, callee.Name) && !g.mayResolveStaticUFCSCall(ctx, call, callee.Name) {
				ctx.setReason(fmt.Sprintf("unresolved static call (%s)", callee.Name))
				return "", "", false
			}
			return g.compileDynamicCall(ctx, call, expected, callee.Name, callNode)
		}
		return g.compileDynamicCall(ctx, call, expected, "", callNode)
	}
	if callee, ok := call.Callee.(*ast.Identifier); ok && callee != nil {
		if info, ok := ctx.functions[callee.Name]; ok && info != nil && info.Compileable {
			needsRuntimeValue := expected == "runtime.Value" && info.ReturnType != "runtime.Value"
			needsExpect := expected != "" && expected != "runtime.Value" && info.ReturnType == "runtime.Value"
			if !g.typeMatches(expected, info.ReturnType) && !needsRuntimeValue && !needsExpect {
				ctx.setReason("call return type mismatch")
				return "", "", false
			}
			optionalLast := g.hasOptionalLastParam(info)
			if len(call.Arguments) != len(info.Params) {
				if !(optionalLast && len(call.Arguments) == len(info.Params)-1) {
					ctx.setReason("call arity mismatch")
					return "", "", false
				}
			}
			missingOptional := optionalLast && len(call.Arguments) == len(info.Params)-1
			if missingOptional && len(info.Params) > 0 && info.Params[len(info.Params)-1].GoType != "runtime.Value" {
				ctx.setReason("call arity mismatch")
				return "", "", false
			}
			args := make([]string, 0, len(call.Arguments))
			preLines := make([]string, 0, len(call.Arguments))
			postLines := make([]string, 0, len(call.Arguments))
			for idx, arg := range call.Arguments {
				param := info.Params[idx]
				if g.typeCategory(param.GoType) == "struct" {
					if ident, ok := arg.(*ast.Identifier); ok && ident != nil {
						if binding, ok := ctx.lookup(ident.Name); ok && binding.GoType == "runtime.Value" {
							runtimeTemp := ctx.newTemp()
							preLines = append(preLines, fmt.Sprintf("%s := %s", runtimeTemp, binding.GoName))
							structExpr, ok := g.expectRuntimeValueExpr(runtimeTemp, param.GoType)
							if !ok {
								ctx.setReason("call argument unsupported")
								return "", "", false
							}
							structTemp := ctx.newTemp()
							preLines = append(preLines, fmt.Sprintf("%s := %s", structTemp, structExpr))
							args = append(args, structTemp)
							baseName, ok := g.structBaseName(param.GoType)
							if !ok {
								baseName = strings.TrimPrefix(param.GoType, "*")
							}
							postLines = append(postLines, fmt.Sprintf("if err := __able_struct_%s_apply(__able_runtime, %s, %s); err != nil { panic(err) }", baseName, runtimeTemp, structTemp))
							continue
						}
					}
				}
				expr, exprType, ok := g.compileExpr(ctx, arg, param.GoType)
				if !ok {
					return "", "", false
				}
				argExpr := expr
				argType := exprType
				if ifaceType, ok := g.interfaceTypeExpr(param.TypeExpr); ok {
					if argType != "runtime.Value" {
						valueExpr, ok := g.runtimeValueExpr(argExpr, argType)
						if !ok {
							ctx.setReason("interface argument unsupported")
							return "", "", false
						}
						argExpr = valueExpr
						argType = "runtime.Value"
					}
					coerced, ok := g.interfaceArgExpr(argExpr, ifaceType, callee.Name, ctx.genericNames)
					if !ok {
						ctx.setReason("interface argument unsupported")
						return "", "", false
					}
					argExpr = coerced
				}
				args = append(args, argExpr)
			}
			if missingOptional {
				args = append(args, "runtime.NilValue{}")
			}
			callExpr := fmt.Sprintf("__able_compiled_%s(%s)", info.GoName, strings.Join(args, ", "))
			bodyLines := []string{
				fmt.Sprintf("__able_push_call_frame(%s)", callNode),
				"defer __able_pop_call_frame()",
			}
			bodyLines = append(bodyLines, preLines...)
			if len(postLines) == 0 {
				bodyLines = append(bodyLines, fmt.Sprintf("return %s", callExpr))
			} else {
				resultTemp := ctx.newTemp()
				bodyLines = append(bodyLines, fmt.Sprintf("%s := %s", resultTemp, callExpr))
				bodyLines = append(bodyLines, postLines...)
				bodyLines = append(bodyLines, fmt.Sprintf("return %s", resultTemp))
			}
			wrapped := fmt.Sprintf("func() %s { %s }()", info.ReturnType, strings.Join(bodyLines, "; "))
			if needsRuntimeValue {
				converted, ok := g.runtimeValueExpr(wrapped, info.ReturnType)
				if !ok {
					ctx.setReason("call return type mismatch")
					return "", "", false
				}
				return converted, "runtime.Value", true
			}
			if needsExpect {
				converted, ok := g.expectRuntimeValueExpr(wrapped, expected)
				if !ok {
					ctx.setReason("call return type mismatch")
					return "", "", false
				}
				return converted, expected, true
			}
			return wrapped, info.ReturnType, true
		}
		if _, ok := ctx.overloads[callee.Name]; ok {
			return g.compileOverloadCall(ctx, call, expected, callee.Name, callNode)
		}
		if _, ok := ctx.lookup(callee.Name); !ok {
			if !g.hasDynamicFeature && !g.mayResolveStaticNamedCall(ctx, callee.Name) && !g.mayResolveStaticUFCSCall(ctx, call, callee.Name) {
				ctx.setReason(fmt.Sprintf("unresolved static call (%s)", callee.Name))
				return "", "", false
			}
			return g.compileDynamicCall(ctx, call, expected, callee.Name, callNode)
		}
	}
	return g.compileDynamicCall(ctx, call, expected, "", callNode)
}

func (g *generator) compileDynamicCall(ctx *compileContext, call *ast.FunctionCall, expected string, calleeName string, callNode string) (string, string, bool) {
	if call == nil {
		ctx.setReason("missing function call")
		return "", "", false
	}
	if expected != "" && expected != "runtime.Value" && !g.isVoidType(expected) && g.typeCategory(expected) == "unknown" {
		ctx.setReason("call return type mismatch")
		return "", "", false
	}
	lines := make([]string, 0, len(call.Arguments)+2)
	args := make([]string, 0, len(call.Arguments))
	calleeTemp := ""
	writebackNeeded := false
	writebackObjExpr := ""
	writebackObjType := ""
	writebackObjTemp := ""
	if calleeName == "" {
		switch callee := call.Callee.(type) {
		case *ast.MemberAccessExpression:
			if callee.Member != nil && !callee.Safe {
				if ident, ok := callee.Member.(*ast.Identifier); ok && ident != nil && ident.Name != "" {
					if method, ok := g.resolveStaticMethodCall(ctx, callee.Object, ident.Name); ok {
						return g.compileResolvedMethodCall(ctx, call, expected, method, "", callNode)
					}
				}
			}
			objExpr, objType, ok := g.compileExpr(ctx, callee.Object, "")
			if !ok {
				return "", "", false
			}
			if callee.Member != nil && !callee.Safe {
				if ident, ok := callee.Member.(*ast.Identifier); ok && ident != nil && ident.Name != "" {
					if expr, retType, ok := g.compileArrayMethodIntrinsicCall(ctx, callee.Object, objExpr, objType, ident.Name, call.Arguments, expected, callNode); ok {
						return expr, retType, true
					}
					if method := g.methodForReceiver(objType, ident.Name); method != nil {
						return g.compileResolvedMethodCall(ctx, call, expected, method, objExpr, callNode)
					}
				}
			}
			if callee.Safe && g.typeCategory(objType) == "runtime" {
				return g.compileSafeMemberCall(ctx, call, callee, expected, objExpr, objType, callNode)
			}
			// Check for impl sibling methods: default interface methods calling
			// sibling methods on self (e.g., describe() calling self.name())
			siblingHandled := false
			if len(ctx.implSiblings) > 0 && ctx.hasImplicitReceiver && !callee.Safe {
				if ident, ok := callee.Member.(*ast.Identifier); ok && ident != nil {
					if sibling, hasSibling := ctx.implSiblings[ident.Name]; hasSibling {
						if objIdent, ok := callee.Object.(*ast.Identifier); ok && objIdent != nil && objIdent.Name == ctx.implicitReceiver.Name {
							objValue, ok := g.runtimeValueExpr(objExpr, objType)
							if ok {
								objTemp := ctx.newTemp()
								calleeTemp = ctx.newTemp()
								lines = append(lines, fmt.Sprintf("%s := %s", objTemp, objValue))
								lines = append(lines, fmt.Sprintf("%s := __able_impl_self_method(%s, %q, %d, __able_wrap_%s)", calleeTemp, objTemp, ident.Name, sibling.Arity, sibling.GoName))
								if g.typeCategory(objType) == "struct" && g.isAddressableMemberObject(callee.Object) {
									writebackNeeded = true
									writebackObjExpr = objExpr
									writebackObjType = objType
									writebackObjTemp = objTemp
								}
								siblingHandled = true
							}
						}
					}
				}
			}
			if !siblingHandled {
				objValue, ok := g.runtimeValueExpr(objExpr, objType)
				if !ok {
					ctx.setReason("method call receiver unsupported")
					return "", "", false
				}
				memberValue, ok := g.memberAssignmentRuntimeValue(ctx, callee.Member)
				if !ok {
					ctx.setReason("method call target unsupported")
					return "", "", false
				}
				objTemp := ctx.newTemp()
				memberTemp := ctx.newTemp()
				calleeTemp = ctx.newTemp()
				lines = append(lines, fmt.Sprintf("%s := %s", objTemp, objValue))
				lines = append(lines, fmt.Sprintf("%s := %s", memberTemp, memberValue))
				lines = append(lines, fmt.Sprintf("%s := __able_member_get_method(%s, %s)", calleeTemp, objTemp, memberTemp))
				if g.typeCategory(objType) == "struct" && g.isAddressableMemberObject(callee.Object) {
					writebackNeeded = true
					writebackObjExpr = objExpr
					writebackObjType = objType
					writebackObjTemp = objTemp
				}
			}
		default:
			calleeExpr, calleeType, ok := g.compileExpr(ctx, call.Callee, "")
			if !ok {
				return "", "", false
			}
			calleeValue, ok := g.runtimeValueExpr(calleeExpr, calleeType)
			if !ok {
				ctx.setReason("call target unsupported")
				return "", "", false
			}
			calleeTemp = ctx.newTemp()
			lines = append(lines, fmt.Sprintf("%s := %s", calleeTemp, calleeValue))
		}
	}

	for _, arg := range call.Arguments {
		expr, goType, ok := g.compileExpr(ctx, arg, "")
		if !ok {
			return "", "", false
		}
		valueExpr, ok := g.runtimeValueExpr(expr, goType)
		if !ok {
			ctx.setReason("call argument unsupported")
			return "", "", false
		}
		temp := ctx.newTemp()
		lines = append(lines, fmt.Sprintf("%s := %s", temp, valueExpr))
		args = append(args, temp)
	}

	argList := strings.Join(args, ", ")
	if argList != "" {
		argList = "[]runtime.Value{" + argList + "}"
	} else {
		argList = "nil"
	}

	callExpr := ""
	if calleeName != "" {
		if wrapper, ok := g.externCallWrapper(calleeName); ok {
			callExpr = fmt.Sprintf("%s(%s, %s)", wrapper, argList, callNode)
		} else {
			callExpr = fmt.Sprintf("__able_call_named(%q, %s, %s)", calleeName, argList, callNode)
		}
	} else {
		callExpr = fmt.Sprintf("__able_call_value(%s, %s, %s)", calleeTemp, argList, callNode)
	}

	if writebackNeeded {
		baseName, ok := g.structBaseName(writebackObjType)
		if !ok {
			baseName = strings.TrimPrefix(writebackObjType, "*")
		}
		convertedTemp := ctx.newTemp()
		writebackLines := []string{
			fmt.Sprintf("%s, err := __able_struct_%s_from(%s)", convertedTemp, baseName, writebackObjTemp),
			"if err != nil { panic(err) }",
		}
		if strings.HasPrefix(writebackObjType, "*") {
			writebackLines = append(writebackLines, fmt.Sprintf("*%s = *%s", writebackObjExpr, convertedTemp))
		} else {
			writebackLines = append(writebackLines, fmt.Sprintf("%s = *%s", writebackObjExpr, convertedTemp))
		}
		if g.isVoidType(expected) {
			lines = append(lines, fmt.Sprintf("_ = %s", callExpr))
			lines = append(lines, writebackLines...)
			return fmt.Sprintf("func() struct{} { %s; return struct{}{} }()", strings.Join(lines, "; ")), "struct{}", true
		}
		resultTemp := ctx.newTemp()
		lines = append(lines, fmt.Sprintf("%s := %s", resultTemp, callExpr))
		lines = append(lines, writebackLines...)
		resultExpr := resultTemp
		resultType := "runtime.Value"
		if expected != "" && expected != "runtime.Value" {
			converted, ok := g.expectRuntimeValueExpr(resultTemp, expected)
			if !ok {
				ctx.setReason("call return type mismatch")
				return "", "", false
			}
			resultExpr = converted
			resultType = expected
		}
		return fmt.Sprintf("func() %s { %s; return %s }()", resultType, strings.Join(lines, "; "), resultExpr), resultType, true
	}

	if g.isVoidType(expected) {
		lines = append(lines, fmt.Sprintf("_ = %s", callExpr))
		return fmt.Sprintf("func() struct{} { %s; return struct{}{} }()", strings.Join(lines, "; ")), "struct{}", true
	}

	resultExpr := callExpr
	resultType := "runtime.Value"
	if expected != "" && expected != "runtime.Value" {
		converted, ok := g.expectRuntimeValueExpr(callExpr, expected)
		if !ok {
			ctx.setReason("call return type mismatch")
			return "", "", false
		}
		resultExpr = converted
		resultType = expected
	}
	if len(lines) == 0 {
		return resultExpr, resultType, true
	}
	return fmt.Sprintf("func() %s { %s; return %s }()", resultType, strings.Join(lines, "; "), resultExpr), resultType, true
}

func (g *generator) externCallWrapper(name string) (string, bool) {
	switch name {
	case "__able_array_new":
		return "__able_extern_array_new", true
	case "__able_array_with_capacity":
		return "__able_extern_array_with_capacity", true
	case "__able_array_size":
		return "__able_extern_array_size", true
	case "__able_array_capacity":
		return "__able_extern_array_capacity", true
	case "__able_array_set_len":
		return "__able_extern_array_set_len", true
	case "__able_array_read":
		return "__able_extern_array_read", true
	case "__able_array_write":
		return "__able_extern_array_write", true
	case "__able_array_reserve":
		return "__able_extern_array_reserve", true
	case "__able_array_clone":
		return "__able_extern_array_clone", true
	case "__able_hash_map_new":
		return "__able_extern_hash_map_new", true
	case "__able_hash_map_with_capacity":
		return "__able_extern_hash_map_with_capacity", true
	case "__able_hash_map_get":
		return "__able_extern_hash_map_get", true
	case "__able_hash_map_set":
		return "__able_extern_hash_map_set", true
	case "__able_hash_map_remove":
		return "__able_extern_hash_map_remove", true
	case "__able_hash_map_contains":
		return "__able_extern_hash_map_contains", true
	case "__able_hash_map_size":
		return "__able_extern_hash_map_size", true
	case "__able_hash_map_clear":
		return "__able_extern_hash_map_clear", true
	case "__able_hash_map_for_each":
		return "__able_extern_hash_map_for_each", true
	case "__able_hash_map_clone":
		return "__able_extern_hash_map_clone", true
	case "__able_String_from_builtin":
		return "__able_extern_string_from_builtin", true
	case "__able_String_to_builtin":
		return "__able_extern_string_to_builtin", true
	case "__able_char_from_codepoint":
		return "__able_extern_char_from_codepoint", true
	case "__able_char_to_codepoint":
		return "__able_extern_char_to_codepoint", true
	case "__able_ratio_from_float":
		return "__able_extern_ratio_from_float", true
	case "__able_f32_bits":
		return "__able_extern_f32_bits", true
	case "__able_f64_bits":
		return "__able_extern_f64_bits", true
	case "__able_u64_mul":
		return "__able_extern_u64_mul", true
	case "__able_channel_new":
		return "__able_extern_channel_new", true
	case "__able_channel_send":
		return "__able_extern_channel_send", true
	case "__able_channel_receive":
		return "__able_extern_channel_receive", true
	case "__able_channel_try_send":
		return "__able_extern_channel_try_send", true
	case "__able_channel_try_receive":
		return "__able_extern_channel_try_receive", true
	case "__able_channel_await_try_recv":
		return "__able_extern_channel_await_try_recv", true
	case "__able_channel_await_try_send":
		return "__able_extern_channel_await_try_send", true
	case "__able_channel_close":
		return "__able_extern_channel_close", true
	case "__able_channel_is_closed":
		return "__able_extern_channel_is_closed", true
	case "__able_mutex_new":
		return "__able_extern_mutex_new", true
	case "__able_mutex_lock":
		return "__able_extern_mutex_lock", true
	case "__able_mutex_unlock":
		return "__able_extern_mutex_unlock", true
	case "__able_mutex_await_lock":
		return "__able_extern_mutex_await_lock", true
	default:
		return "", false
	}
}

func (g *generator) resolveStaticMethodCall(ctx *compileContext, object ast.Expression, memberName string) (*methodInfo, bool) {
	if g == nil || object == nil || memberName == "" {
		return nil, false
	}
	ident, ok := object.(*ast.Identifier)
	if !ok || ident == nil || ident.Name == "" {
		return nil, false
	}
	if ctx != nil {
		if _, ok := ctx.lookup(ident.Name); ok {
			return nil, false
		}
	}
	if _, ok := g.structInfoForTypeName(ctx.packageName, ident.Name); !ok {
		return nil, false
	}
	method := g.methodForTypeNameInPackage(ctx.packageName, ident.Name, memberName, false)
	if method == nil {
		return nil, false
	}
	return method, true
}

func (g *generator) compileResolvedMethodCall(ctx *compileContext, call *ast.FunctionCall, expected string, method *methodInfo, receiverExpr string, callNode string) (string, string, bool) {
	if call == nil || method == nil || method.Info == nil {
		ctx.setReason("missing method call")
		return "", "", false
	}
	info := method.Info
	if !info.Compileable {
		ctx.setReason("unsupported method call")
		return "", "", false
	}
	if !g.typeMatches(expected, info.ReturnType) {
		ctx.setReason("call return type mismatch")
		return "", "", false
	}
	paramOffset := 0
	args := make([]string, 0, len(call.Arguments)+1)
	if method.ExpectsSelf {
		if receiverExpr == "" {
			ctx.setReason("method receiver missing")
			return "", "", false
		}
		args = append(args, receiverExpr)
		paramOffset = 1
	}
	params := info.Params
	if paramOffset > len(params) {
		ctx.setReason("method params missing")
		return "", "", false
	}
	callArgCount := len(call.Arguments)
	paramCount := len(params) - paramOffset
	optionalLast := g.hasOptionalLastParam(info)
	if callArgCount != paramCount {
		if !(optionalLast && callArgCount == paramCount-1) {
			ctx.setReason("call arity mismatch")
			return "", "", false
		}
	}
	missingOptional := optionalLast && callArgCount == paramCount-1
	if missingOptional && paramCount > 0 && params[len(params)-1].GoType != "runtime.Value" {
		ctx.setReason("call arity mismatch")
		return "", "", false
	}
	for idx, arg := range call.Arguments {
		param := params[paramOffset+idx]
		expr, exprType, ok := g.compileExpr(ctx, arg, param.GoType)
		if !ok {
			return "", "", false
		}
		argExpr := expr
		argType := exprType
		if ifaceType, ok := g.interfaceTypeExpr(param.TypeExpr); ok {
			if argType != "runtime.Value" {
				valueExpr, ok := g.runtimeValueExpr(argExpr, argType)
				if !ok {
					ctx.setReason("interface argument unsupported")
					return "", "", false
				}
				argExpr = valueExpr
				argType = "runtime.Value"
			}
			coerced, ok := g.interfaceArgExpr(argExpr, ifaceType, info.Name, ctx.genericNames)
			if !ok {
				ctx.setReason("interface argument unsupported")
				return "", "", false
			}
			argExpr = coerced
		}
		args = append(args, argExpr)
	}
	if missingOptional {
		args = append(args, "runtime.NilValue{}")
	}
	callExpr := fmt.Sprintf("__able_compiled_%s(%s)", info.GoName, strings.Join(args, ", "))
	return fmt.Sprintf("func() %s { __able_push_call_frame(%s); defer __able_pop_call_frame(); return %s }()", info.ReturnType, callNode, callExpr), info.ReturnType, true
}

func (g *generator) compileSafeMemberCall(ctx *compileContext, call *ast.FunctionCall, callee *ast.MemberAccessExpression, expected string, objExpr string, objType string, callNode string) (string, string, bool) {
	if call == nil || callee == nil {
		ctx.setReason("missing safe member call")
		return "", "", false
	}
	objValue, ok := g.runtimeValueExpr(objExpr, objType)
	if !ok {
		ctx.setReason("method call receiver unsupported")
		return "", "", false
	}
	memberValue, ok := g.memberAssignmentRuntimeValue(ctx, callee.Member)
	if !ok {
		ctx.setReason("method call target unsupported")
		return "", "", false
	}
	lines := make([]string, 0, len(call.Arguments)+4)
	args := make([]string, 0, len(call.Arguments))
	objTemp := ctx.newTemp()
	memberTemp := ctx.newTemp()
	calleeTemp := ctx.newTemp()
	lines = append(lines, fmt.Sprintf("%s := %s", objTemp, objValue))
	lines = append(lines, fmt.Sprintf("if __able_is_nil(%s) { return %s }", objTemp, safeNilReturnExpr(expected)))
	lines = append(lines, fmt.Sprintf("%s := %s", memberTemp, memberValue))
	lines = append(lines, fmt.Sprintf("%s := __able_member_get_method(%s, %s)", calleeTemp, objTemp, memberTemp))
	for _, arg := range call.Arguments {
		expr, goType, ok := g.compileExpr(ctx, arg, "")
		if !ok {
			return "", "", false
		}
		valueExpr, ok := g.runtimeValueExpr(expr, goType)
		if !ok {
			ctx.setReason("call argument unsupported")
			return "", "", false
		}
		temp := ctx.newTemp()
		lines = append(lines, fmt.Sprintf("%s := %s", temp, valueExpr))
		args = append(args, temp)
	}
	argList := strings.Join(args, ", ")
	if argList != "" {
		argList = "[]runtime.Value{" + argList + "}"
	} else {
		argList = "nil"
	}
	callExpr := fmt.Sprintf("__able_call_value(%s, %s, %s)", calleeTemp, argList, callNode)
	if g.isVoidType(expected) {
		lines = append(lines, fmt.Sprintf("_ = %s", callExpr))
		lines = append(lines, "return struct{}{}")
		return fmt.Sprintf("func() struct{} { %s }()", strings.Join(lines, "; ")), "struct{}", true
	}
	lines = append(lines, fmt.Sprintf("return %s", callExpr))
	baseExpr := fmt.Sprintf("func() runtime.Value { %s }()", strings.Join(lines, "; "))
	if expected == "" || expected == "runtime.Value" {
		return baseExpr, "runtime.Value", true
	}
	converted, ok := g.expectRuntimeValueExpr(baseExpr, expected)
	if !ok {
		ctx.setReason("call return type mismatch")
		return "", "", false
	}
	return converted, expected, true
}

func safeNilReturnExpr(expected string) string {
	if expected == "struct{}" {
		return "struct{}{}"
	}
	return "runtime.NilValue{}"
}
