package compiler

import (
	"fmt"
	"strings"

	"able/interpreter-go/pkg/ast"
)

func (g *generator) compileFunctionCall(ctx *compileContext, call *ast.FunctionCall, expected string) ([]string, string, string, bool) {
	if call == nil {
		ctx.setReason("missing function call")
		return nil, "", "", false
	}
	callNode := g.diagNodeName(call, "*ast.FunctionCall", "call")
	if callee, ok := call.Callee.(*ast.MemberAccessExpression); ok && callee != nil && !callee.Safe && callee.Member != nil {
		if ident, ok := callee.Member.(*ast.Identifier); ok && ident != nil && ident.Name != "" {
			if typeIdent, ok := callee.Object.(*ast.Identifier); ok && typeIdent != nil && typeIdent.Name != "" {
				if lines, expr, goType, ok := g.compileStaticArrayFactoryCall(ctx, typeIdent.Name, ident.Name, call.Arguments, expected, callNode); ok {
					return lines, expr, goType, true
				}
			}
		}
	}
	if len(call.TypeArguments) > 0 {
		if callee, ok := call.Callee.(*ast.MemberAccessExpression); ok && callee != nil && !callee.Safe && callee.Member != nil {
			if ident, ok := callee.Member.(*ast.Identifier); ok && ident != nil && ident.Name != "" {
				objLines, objExpr, objType, ok := g.compileExprLines(ctx, callee.Object, "")
				if ok {
					if siblingLines, expr, retType, ok := g.compileDirectImplSiblingMethodCall(ctx, call, callee, expected, objExpr, objType, callNode); ok {
						lines := append(objLines, siblingLines...)
						return lines, expr, retType, true
					}
					if method := g.methodForReceiver(objType, ident.Name); method != nil {
						method = g.concreteMethodCallInfo(ctx, call, method, callee.Object, objType, expected)
						methodLines, expr, retType, ok := g.compileResolvedMethodCall(ctx, call, expected, method, objExpr, objType, callNode)
						if ok {
							lines := append(objLines, methodLines...)
							return lines, expr, retType, true
						}
					}
					if method := g.compileableInterfaceMethodForConcreteReceiver(objType, ident.Name); method != nil {
						method = g.concreteMethodCallInfo(ctx, call, method, callee.Object, objType, expected)
						methodLines, expr, retType, ok := g.compileResolvedMethodCall(ctx, call, expected, method, objExpr, objType, callNode)
						if ok {
							lines := append(objLines, methodLines...)
							return lines, expr, retType, true
						}
					}
					if ifaceLines, expr, retType, ok := g.compileNativeInterfaceGenericMethodCall(ctx, call, expected, objExpr, objType, ident.Name, callNode); ok {
						lines := append(objLines, ifaceLines...)
						return lines, expr, retType, true
					}
				}
			}
		}
		if callee, ok := call.Callee.(*ast.Identifier); ok && callee != nil {
			// Generic calls on local values (e.g. generic lambdas) must call the
			// bound value, not global name lookup.
			if binding, ok := ctx.lookup(callee.Name); ok {
				if isFunctionTypeExpr(binding.TypeExpr) {
					return g.compileFnParamCall(ctx, call, expected, binding, callNode)
				}
				return g.compileDynamicCall(ctx, call, expected, "", callNode)
			}
			if !g.hasDynamicFeature && !g.mayResolveStaticNamedCall(ctx, callee.Name) && !g.mayResolveStaticUFCSCall(ctx, call, callee.Name) {
				ctx.setReason(fmt.Sprintf("unresolved static call (%s)", callee.Name))
				return nil, "", "", false
			}
			return g.compileDynamicCall(ctx, call, expected, callee.Name, callNode)
		}
		return g.compileDynamicCall(ctx, call, expected, "", callNode)
	}
	if callee, ok := call.Callee.(*ast.Identifier); ok && callee != nil {
		if info, overload, ok := g.resolveStaticCallable(ctx, callee.Name); ok {
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
				var writebackIdents []string // Able variable names that get struct writeback
				for idx, arg := range call.Arguments {
					param := info.Params[idx]
					if g.typeCategory(param.GoType) == "struct" {
						if ident, ok := arg.(*ast.Identifier); ok && ident != nil {
							if binding, ok := ctx.lookup(ident.Name); ok && binding.GoType == "runtime.Value" {
								runtimeTemp := ctx.newTemp()
								preLines = append(preLines, fmt.Sprintf("%s := %s", runtimeTemp, binding.GoName))
								convLines, structExpr, ok := g.expectRuntimeValueExprLines(ctx, runtimeTemp, param.GoType)
								if !ok {
									ctx.setReason("call argument unsupported")
									return nil, "", "", false
								}
								preLines = append(preLines, convLines...)
								structTemp := ctx.newTemp()
								preLines = append(preLines, fmt.Sprintf("%s := %s", structTemp, structExpr))
								args = append(args, structTemp)
								baseName, ok := g.structBaseName(param.GoType)
								if !ok {
									baseName = strings.TrimPrefix(param.GoType, "*")
								}
								transferLines, ok := g.controlTransferLines(ctx, g.runtimeErrorControlExpr(callNode, "err"))
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
					argLines, expr, exprType, ok := g.compileExprLinesWithExpectedTypeExpr(ctx, arg, param.GoType, param.TypeExpr)
					if !ok {
						return nil, "", "", false
					}
					preLines = append(preLines, argLines...)
					argExpr := expr
					argType := exprType
					if ifaceType, ok := g.interfaceTypeExpr(param.TypeExpr); ok && param.GoType == "runtime.Value" {
						if argType != "runtime.Value" {
							argConvLines, valueExpr, ok := g.runtimeValueLines(ctx, argExpr, argType)
							if !ok {
								ctx.setReason("interface argument unsupported")
								return nil, "", "", false
							}
							preLines = append(preLines, argConvLines...)
							argExpr = valueExpr
							argType = "runtime.Value"
						}
						ifaceLines, coerced, ok := g.interfaceArgExprLines(ctx, argExpr, ifaceType, callee.Name, ctx.genericNames)
						if !ok {
							ctx.setReason("interface argument unsupported")
							return nil, "", "", false
						}
						preLines = append(preLines, ifaceLines...)
						argExpr = coerced
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
				controlLines, ok := g.controlCheckLines(ctx, controlTemp)
				if !ok {
					return nil, "", "", false
				}
				lines = append(lines, controlLines...)
				lines = append(lines, postLines...)
				// Writeback invalidates cached struct extractions for written-back variables.
				if len(writebackIdents) > 0 && ctx.originExtractions != nil {
					for _, name := range writebackIdents {
						delete(ctx.originExtractions, name)
					}
				}
				if needsRuntimeValue {
					convLines, converted, ok := g.runtimeValueLines(ctx, resultTemp, info.ReturnType)
					if !ok {
						ctx.setReason("call return type mismatch")
						return nil, "", "", false
					}
					lines = append(lines, convLines...)
					return lines, converted, "runtime.Value", true
				}
				if needsExpect {
					convLines, converted, ok := g.expectRuntimeValueExprLines(ctx, resultTemp, expected)
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
					convLines, converted, ok := g.expectRuntimeValueExprLines(ctx, anyTemp, expected)
					if !ok {
						ctx.setReason("call return type mismatch")
						return nil, "", "", false
					}
					lines = append(lines, convLines...)
					return lines, converted, expected, true
				}
				if needsStaticCoerce {
					return g.coerceExpectedStaticExpr(ctx, lines, resultTemp, info.ReturnType, expected)
				}
				return lines, resultTemp, info.ReturnType, true
			}
			if overload != nil {
				return g.compileResolvedOverloadCall(ctx, call, expected, overload.Package, overload.Name, callNode)
			}
		}
		binding, found := ctx.lookup(callee.Name)
		if !found {
			if !g.hasDynamicFeature && !g.mayResolveStaticNamedCall(ctx, callee.Name) && !g.mayResolveStaticUFCSCall(ctx, call, callee.Name) {
				ctx.setReason(fmt.Sprintf("unresolved static call (%s)", callee.Name))
				return nil, "", "", false
			}
			return g.compileDynamicCall(ctx, call, expected, callee.Name, callNode)
		}
		// When calling a local binding whose type is a function (e.g. a lambda
		// parameter like `predicate: T -> bool`), use the fast call path which
		// avoids interface unwrapping, partial application checks, and overload
		// resolution overhead.
		if isFunctionTypeExpr(binding.TypeExpr) || g.nativeCallableInfoForGoType(binding.GoType) != nil {
			return g.compileFnParamCall(ctx, call, expected, binding, callNode)
		}
	}
	return g.compileDynamicCall(ctx, call, expected, "", callNode)
}

// compileFnParamCall compiles a call to a local binding known to be a function
// value (e.g. a lambda parameter like `predicate: T -> bool`). Uses the fast
// call path that type-switches directly on NativeFunctionValue, skipping
// interface unwrapping, partial application, thunk, and overload resolution.
func (g *generator) compileFnParamCall(ctx *compileContext, call *ast.FunctionCall, expected string, binding paramInfo, callNode string) ([]string, string, string, bool) {
	if info := g.nativeCallableInfoForGoType(binding.GoType); info != nil {
		var fnTypeExpr *ast.FunctionTypeExpression
		if typed, ok := binding.TypeExpr.(*ast.FunctionTypeExpression); ok {
			fnTypeExpr = typed
		}
		return g.compileNativeCallableCall(ctx, call, expected, binding.GoName, info.GoType, fnTypeExpr, callNode)
	}
	lines := make([]string, 0, len(call.Arguments)+4)
	args := make([]string, 0, len(call.Arguments))
	for _, arg := range call.Arguments {
		argLines, expr, goType, ok := g.compileExprLines(ctx, arg, "")
		if !ok {
			return nil, "", "", false
		}
		lines = append(lines, argLines...)
		argConvLines, valueExpr, ok := g.runtimeValueLines(ctx, expr, goType)
		if !ok {
			ctx.setReason("call argument unsupported")
			return nil, "", "", false
		}
		lines = append(lines, argConvLines...)
		temp := ctx.newTemp()
		lines = append(lines, fmt.Sprintf("%s := %s", temp, valueExpr))
		args = append(args, temp)
	}
	argList := "nil"
	if len(args) > 0 {
		argList = "[]runtime.Value{" + strings.Join(args, ", ") + "}"
	}
	resultTemp := ctx.newTemp()
	errTemp := ctx.newTemp()
	controlTemp := ctx.newTemp()
	callableExpr := binding.GoName
	if binding.GoType == "any" {
		callableTemp := ctx.newTemp()
		lines = append(lines, fmt.Sprintf("%s := __able_any_to_value(%s)", callableTemp, binding.GoName))
		callableExpr = callableTemp
	}
	lines = append(lines,
		fmt.Sprintf("%s, %s := __able_call_value_fast(%s, %s)", resultTemp, errTemp, callableExpr, argList),
		fmt.Sprintf("%s := __able_control_from_error(%s)", controlTemp, errTemp),
	)
	controlLines, ok := g.controlCheckLines(ctx, controlTemp)
	if !ok {
		return nil, "", "", false
	}
	lines = append(lines, controlLines...)
	resultExpr := resultTemp
	resultType := "runtime.Value"
	if g.isVoidType(expected) {
		lines = append(lines, fmt.Sprintf("_ = %s", resultTemp))
		return lines, "struct{}{}", "struct{}", true
	}
	if expected != "" && expected != "runtime.Value" {
		convLines, converted, ok := g.expectRuntimeValueExprLines(ctx, resultTemp, expected)
		if !ok {
			ctx.setReason("call return type mismatch")
			return nil, "", "", false
		}
		lines = append(lines, convLines...)
		resultExpr = converted
		resultType = expected
	}
	return lines, resultExpr, resultType, true
}

func (g *generator) compileDynamicCall(ctx *compileContext, call *ast.FunctionCall, expected string, calleeName string, callNode string) ([]string, string, string, bool) {
	if call == nil {
		ctx.setReason("missing function call")
		return nil, "", "", false
	}
	if !g.isStaticallyKnownExpectedType(expected) {
		ctx.setReason("call return type mismatch")
		return nil, "", "", false
	}
	lines := make([]string, 0, len(call.Arguments)+2)
	args := make([]string, 0, len(call.Arguments))
	calleeTemp := ""
	fastMethodName := ""
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
						method = g.concreteStaticMethodCallInfo(ctx, call, method, callee.Object, expected)
						return g.compileResolvedMethodCall(ctx, call, expected, method, "", "", callNode)
					}
				}
			}
			objLines, objExpr, objType, ok := g.compileExprLines(ctx, callee.Object, "")
			if !ok {
				return nil, "", "", false
			}
			lines = append(lines, objLines...)
			if callee.Member != nil && !callee.Safe {
				if ident, ok := callee.Member.(*ast.Identifier); ok && ident != nil && ident.Name != "" {
					if controllerLines, expr, retType, ok := g.compileStaticIteratorControllerCall(ctx, call, expected, objExpr, objType, ident.Name); ok {
						lines = append(lines, controllerLines...)
						return lines, expr, retType, true
					}
					if intrLines, expr, retType, ok := g.compileArrayMethodIntrinsicCall(ctx, callee.Object, objExpr, objType, ident.Name, call.Arguments, expected, callNode); ok {
						lines = append(lines, intrLines...)
						return lines, expr, retType, true
					}
					if errorLines, expr, retType, ok := g.compileNativeErrorMethodCall(ctx, call, expected, objExpr, objType, ident.Name, callNode); ok {
						lines = append(lines, errorLines...)
						return lines, expr, retType, true
					}
					if ifaceLines, expr, retType, ok := g.compileNativeInterfaceMethodCall(ctx, call, expected, objExpr, objType, ident.Name, callNode); ok {
						lines = append(lines, ifaceLines...)
						return lines, expr, retType, true
					}
					if siblingLines, expr, retType, ok := g.compileDirectImplSiblingMethodCall(ctx, call, callee, expected, objExpr, objType, callNode); ok {
						lines = append(lines, siblingLines...)
						return lines, expr, retType, true
					}
					if method := g.methodForReceiver(objType, ident.Name); method != nil {
						method = g.concreteMethodCallInfo(ctx, call, method, callee.Object, objType, expected)
						methodLines, v, t, ok := g.compileResolvedMethodCall(ctx, call, expected, method, objExpr, objType, callNode)
						lines = append(lines, methodLines...)
						return lines, v, t, ok
					}
					if method := g.compileableInterfaceMethodForConcreteReceiver(objType, ident.Name); method != nil {
						method = g.concreteMethodCallInfo(ctx, call, method, callee.Object, objType, expected)
						methodLines, v, t, ok := g.compileResolvedMethodCall(ctx, call, expected, method, objExpr, objType, callNode)
						lines = append(lines, methodLines...)
						return lines, v, t, ok
					}
					if ifaceLines, expr, retType, ok := g.compileNativeInterfaceGenericMethodCall(ctx, call, expected, objExpr, objType, ident.Name, callNode); ok {
						lines = append(lines, ifaceLines...)
						return lines, expr, retType, true
					}
					// When receiver is runtime.Value with known origin struct type,
					// extract the struct, call the compiled method directly, and writeback.
					if objType == "runtime.Value" {
						if originLines, v, t, ok := g.compileOriginStructMethodCall(ctx, call, expected, callee, objExpr, ident.Name, callNode); ok {
							lines = append(lines, originLines...)
							return lines, v, t, true
						}
					}
				}
			}
			if callee.Safe {
				cat := g.typeCategory(objType)
				if cat == "runtime" || cat == "any" || (cat == "struct" && strings.HasPrefix(objType, "*")) {
					safeLines, v, t, ok := g.compileSafeMemberCall(ctx, call, callee, expected, objExpr, objType, callNode)
					lines = append(lines, safeLines...)
					return lines, v, t, ok
				}
			}
			// Check for impl sibling methods: default interface methods calling
			// sibling methods on self (e.g., describe() calling self.name())
			siblingHandled := false
			if len(ctx.implSiblings) > 0 && ctx.hasImplicitReceiver && !callee.Safe {
				if ident, ok := callee.Member.(*ast.Identifier); ok && ident != nil {
					if sibling, hasSibling := ctx.implSiblings[ident.Name]; hasSibling {
						if objIdent, ok := callee.Object.(*ast.Identifier); ok && objIdent != nil && objIdent.Name == ctx.implicitReceiver.Name {
							// Try direct compiled call first. Receiver coercion is handled
							// by compileResolvedMethodCall, so runtime.Value receivers can
							// still stay on the static path when they have a known native ABI.
							if sibling.Info != nil && sibling.Info.Compileable && len(sibling.Info.Params) > 0 {
								method := &methodInfo{
									MethodName:  ident.Name,
									ExpectsSelf: true,
									Info:        sibling.Info,
								}
								method = g.concreteMethodCallInfo(ctx, call, method, callee.Object, objType, expected)
								methodLines, v, t, ok := g.compileResolvedMethodCall(ctx, call, expected, method, objExpr, objType, callNode)
								if ok {
									lines = append(lines, methodLines...)
									return lines, v, t, true
								}
							}
							// Fall back to dynamic dispatch
							objConvLines, objValue, ok := g.runtimeValueLines(ctx, objExpr, objType)
							if ok {
								objTemp := ctx.newTemp()
								calleeTemp = ctx.newTemp()
								lines = append(lines, objConvLines...)
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
				objConvLines, objValue, ok := g.runtimeValueLines(ctx, objExpr, objType)
				if !ok {
					ctx.setReason("method call receiver unsupported")
					return nil, "", "", false
				}
				objTemp := ctx.newTemp()
				lines = append(lines, objConvLines...)
				lines = append(lines, fmt.Sprintf("%s := %s", objTemp, objValue))
				// When the member is a literal identifier, use __able_method_call
				// which combines member_get_method + call_value_fast in one step,
				// avoiding bridge.ToString allocation and __able_call_value overhead.
				if ident, ok := callee.Member.(*ast.Identifier); ok && ident != nil && ident.Name != "" {
					fastMethodName = ident.Name
					calleeTemp = objTemp
				} else {
					memberValue, ok := g.memberAssignmentRuntimeValue(ctx, callee.Member)
					if !ok {
						ctx.setReason("method call target unsupported")
						return nil, "", "", false
					}
					memberTemp := ctx.newTemp()
					lines = append(lines, fmt.Sprintf("%s := %s", memberTemp, memberValue))
					lines, calleeTemp, ok = g.appendRuntimeMemberGetMethodControlLines(ctx, lines, objTemp, memberTemp)
					if !ok {
						return nil, "", "", false
					}
				}
				if g.typeCategory(objType) == "struct" && g.isAddressableMemberObject(callee.Object) {
					writebackNeeded = true
					writebackObjExpr = objExpr
					writebackObjType = objType
					writebackObjTemp = objTemp
				}
			}
		default:
			calleeLines, calleeExpr, calleeType, ok := g.compileExprLines(ctx, call.Callee, "")
			if !ok {
				return nil, "", "", false
			}
			lines = append(lines, calleeLines...)
			if applyLines, expr, retType, ok := g.compileStaticApplyCall(ctx, call, expected, calleeExpr, calleeType, callNode); ok {
				lines = append(lines, applyLines...)
				return lines, expr, retType, true
			}
			calleeConvLines, calleeValue, ok := g.runtimeValueLines(ctx, calleeExpr, calleeType)
			if !ok {
				ctx.setReason("call target unsupported")
				return nil, "", "", false
			}
			calleeTemp = ctx.newTemp()
			lines = append(lines, calleeConvLines...)
			lines = append(lines, fmt.Sprintf("%s := %s", calleeTemp, calleeValue))
		}
	}

	for _, arg := range call.Arguments {
		argLines, expr, goType, ok := g.compileExprLines(ctx, arg, "")
		if !ok {
			return nil, "", "", false
		}
		lines = append(lines, argLines...)
		argConvLines, valueExpr, ok := g.runtimeValueLines(ctx, expr, goType)
		if !ok {
			ctx.setReason("call argument unsupported")
			return nil, "", "", false
		}
		lines = append(lines, argConvLines...)
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
			var ok bool
			lines, callExpr, ok = g.appendRuntimeCallControlLines(ctx, lines, fmt.Sprintf("__able_call_named(%q, %s, %s)", calleeName, argList, callNode))
			if !ok {
				return nil, "", "", false
			}
		}
	} else if fastMethodName != "" {
		// Fast path: combined member lookup + call.
		resultTemp := ctx.newTemp()
		controlTemp := ctx.newTemp()
		lines = append(lines,
			fmt.Sprintf("%s, %s := __able_method_call_node(%s, %q, %s, %s)", resultTemp, controlTemp, calleeTemp, fastMethodName, argList, callNode),
		)
		controlLines, ok := g.controlCheckLines(ctx, controlTemp)
		if !ok {
			return nil, "", "", false
		}
		lines = append(lines, controlLines...)
		callExpr = resultTemp
	} else {
		var ok bool
		lines, callExpr, ok = g.appendRuntimeCallControlLines(ctx, lines, fmt.Sprintf("__able_call_value(%s, %s, %s)", calleeTemp, argList, callNode))
		if !ok {
			return nil, "", "", false
		}
	}

	if writebackNeeded {
		baseName, ok := g.structBaseName(writebackObjType)
		if !ok {
			baseName = strings.TrimPrefix(writebackObjType, "*")
		}
		resultTemp := ctx.newTemp()
		lines = append(lines, fmt.Sprintf("%s := %s", resultTemp, callExpr))
		convertedTemp := ctx.newTemp()
		errTemp := ctx.newTemp()
		lines = append(lines,
			fmt.Sprintf("%s, %s := __able_struct_%s_from(%s)", convertedTemp, errTemp, baseName, writebackObjTemp),
		)
		controlTemp := ctx.newTemp()
		lines = append(lines, fmt.Sprintf("%s := __able_control_from_error(%s)", controlTemp, errTemp))
		controlLines, ok := g.controlCheckLines(ctx, controlTemp)
		if !ok {
			return nil, "", "", false
		}
		lines = append(lines, controlLines...)
		if strings.HasPrefix(writebackObjType, "*") {
			lines = append(lines, fmt.Sprintf("*%s = *%s", writebackObjExpr, convertedTemp))
		} else {
			lines = append(lines, fmt.Sprintf("%s = *%s", writebackObjExpr, convertedTemp))
		}
		if g.isVoidType(expected) {
			return lines, "struct{}{}", "struct{}", true
		}
		resultExpr := resultTemp
		resultType := "runtime.Value"
		if expected != "" && expected != "runtime.Value" {
			convLines, converted, ok := g.expectRuntimeValueExprLines(ctx, resultTemp, expected)
			if !ok {
				ctx.setReason("call return type mismatch")
				return nil, "", "", false
			}
			lines = append(lines, convLines...)
			resultExpr = converted
			resultType = expected
		}
		return lines, resultExpr, resultType, true
	}

	if g.isVoidType(expected) {
		lines = append(lines, fmt.Sprintf("_ = %s", callExpr))
		return lines, "struct{}{}", "struct{}", true
	}

	resultExpr := callExpr
	resultType := "runtime.Value"
	if expected != "" && expected != "runtime.Value" {
		convLines, converted, ok := g.expectRuntimeValueExprLines(ctx, callExpr, expected)
		if !ok {
			ctx.setReason("call return type mismatch")
			return nil, "", "", false
		}
		lines = append(lines, convLines...)
		resultExpr = converted
		resultType = expected
	}
	return lines, resultExpr, resultType, true
}

func (g *generator) compileDirectImplSiblingMethodCall(ctx *compileContext, call *ast.FunctionCall, callee *ast.MemberAccessExpression, expected string, objExpr string, objType string, callNode string) ([]string, string, string, bool) {
	if g == nil || ctx == nil || call == nil || callee == nil || callee.Safe || !ctx.hasImplicitReceiver || len(ctx.implSiblings) == 0 {
		return nil, "", "", false
	}
	ident, ok := callee.Member.(*ast.Identifier)
	if !ok || ident == nil || ident.Name == "" {
		return nil, "", "", false
	}
	objIdent, ok := callee.Object.(*ast.Identifier)
	if !ok || objIdent == nil || objIdent.Name != ctx.implicitReceiver.Name {
		return nil, "", "", false
	}
	sibling, ok := ctx.implSiblings[ident.Name]
	if !ok || sibling.Info == nil || !sibling.Info.Compileable || len(sibling.Info.Params) == 0 {
		return nil, "", "", false
	}
	method := &methodInfo{
		MethodName:  ident.Name,
		ExpectsSelf: true,
		Info:        sibling.Info,
	}
	method = g.concreteMethodCallInfo(ctx, call, method, callee.Object, objType, expected)
	return g.compileResolvedMethodCall(ctx, call, expected, method, objExpr, objType, callNode)
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

func (g *generator) compileSafeMemberCall(ctx *compileContext, call *ast.FunctionCall, callee *ast.MemberAccessExpression, expected string, objExpr string, objType string, callNode string) ([]string, string, string, bool) {
	if call == nil || callee == nil {
		ctx.setReason("missing safe member call")
		return nil, "", "", false
	}
	objConvLines, objValue, ok := g.runtimeValueLines(ctx, objExpr, objType)
	if !ok {
		ctx.setReason("method call receiver unsupported")
		return nil, "", "", false
	}
	memberValue, ok := g.memberAssignmentRuntimeValue(ctx, callee.Member)
	if !ok {
		ctx.setReason("method call target unsupported")
		return nil, "", "", false
	}
	// Pre-compile arguments before the if/else so temps are in scope.
	argConvLinesList := make([][]string, 0, len(call.Arguments))
	argTemps := make([]string, 0, len(call.Arguments))
	var argPreLines []string
	for _, arg := range call.Arguments {
		argLines, expr, goType, ok := g.compileExprLines(ctx, arg, "")
		if !ok {
			return nil, "", "", false
		}
		argPreLines = append(argPreLines, argLines...)
		convLines, valueExpr, ok := g.runtimeValueLines(ctx, expr, goType)
		if !ok {
			ctx.setReason("call argument unsupported")
			return nil, "", "", false
		}
		argConvLinesList = append(argConvLinesList, convLines)
		argTemps = append(argTemps, valueExpr)
	}
	resultType := "runtime.Value"
	if g.isVoidType(expected) {
		resultType = "struct{}"
	}
	resultTemp := ctx.newTemp()
	objTemp := ctx.newTemp()
	lines := make([]string, 0, len(argPreLines)+len(call.Arguments)*2+8+len(objConvLines))
	lines = append(lines, objConvLines...)
	lines = append(lines, fmt.Sprintf("%s := %s", objTemp, objValue))
	lines = append(lines, fmt.Sprintf("var %s %s", resultTemp, resultType))
	lines = append(lines, fmt.Sprintf("if __able_is_nil(%s) {", objTemp))
	lines = append(lines, fmt.Sprintf("%s = %s", resultTemp, safeNilReturnExpr(expected)))
	lines = append(lines, "} else {")
	lines = append(lines, indentLines(argPreLines, 1)...)
	memberTemp := ctx.newTemp()
	calleeTemp := ""
	lines = append(lines, fmt.Sprintf("%s := %s", memberTemp, memberValue))
	lines, calleeTemp, ok = g.appendRuntimeMemberGetMethodControlLines(ctx, lines, objTemp, memberTemp)
	if !ok {
		return nil, "", "", false
	}
	args := make([]string, 0, len(call.Arguments))
	for i, valueExpr := range argTemps {
		lines = append(lines, argConvLinesList[i]...)
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
	lines, callExpr, ok = g.appendRuntimeCallControlLines(ctx, lines, fmt.Sprintf("__able_call_value(%s, %s, %s)", calleeTemp, argList, callNode))
	if !ok {
		return nil, "", "", false
	}
	if g.isVoidType(expected) {
		lines = append(lines, fmt.Sprintf("_ = %s", callExpr))
	} else {
		lines = append(lines, fmt.Sprintf("%s = %s", resultTemp, callExpr))
	}
	lines = append(lines, "}")
	if expected != "" && expected != "runtime.Value" && !g.isVoidType(expected) {
		convLines, converted, ok := g.expectRuntimeValueExprLines(ctx, resultTemp, expected)
		if !ok {
			ctx.setReason("call return type mismatch")
			return nil, "", "", false
		}
		lines = append(lines, convLines...)
		return lines, converted, expected, true
	}
	return lines, resultTemp, resultType, true
}
