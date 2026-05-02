package compiler

import (
	"fmt"
	"strings"

	"able/interpreter-go/pkg/ast"
)

type dynamicCallArgExpectation struct {
	goType   string
	typeExpr ast.TypeExpression
}

func (g *generator) compileFunctionCall(ctx *compileContext, call *ast.FunctionCall, expected string) ([]string, string, string, bool) {
	if call == nil {
		ctx.setReason("missing function call")
		return nil, "", "", false
	}
	callNode := g.diagNodeName(call, "*ast.FunctionCall", "call")
	if callee, ok := call.Callee.(*ast.MemberAccessExpression); ok && callee != nil && !callee.Safe {
		if lines, expr, goType, ok := g.compileStaticPackageSelectorCall(ctx, call, expected, callee, callNode); ok {
			return lines, expr, goType, true
		}
	}
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
				preferredReceiverTypeExpr := g.preferredDispatchReceiverTypeExpr(ctx, call, callee.Object, ident.Name, expected)
				objLines, objExpr, objType, ok := g.compileDispatchReceiverExprWithExpectedTypeExpr(ctx, callee.Object, "", preferredReceiverTypeExpr)
				if ok {
					if siblingLines, expr, retType, ok := g.compileDirectImplSiblingMethodCall(ctx, call, callee, expected, objExpr, objType, callNode); ok {
						lines := siblingLines
						if !g.usesImplicitReceiverSiblingCall(ctx, callee.Object) {
							lines = append(objLines, siblingLines...)
						}
						return lines, expr, retType, true
					}
					if method := g.methodForReceiver(objType, ident.Name); method != nil {
						method = g.concreteMethodCallInfo(ctx, call, method, callee.Object, objType, expected)
						methodLines, expr, retType, ok := g.lowerResolvedMethodDispatch(ctx, call, expected, method, objExpr, objType, callNode)
						if ok {
							lines := append(objLines, methodLines...)
							return lines, expr, retType, true
						}
					}
					if method := g.compileableInterfaceMethodForConcreteReceiverExpr(ctx, callee.Object, objType, ident.Name); method != nil {
						method = g.concreteMethodCallInfo(ctx, call, method, callee.Object, objType, expected)
						methodLines, expr, retType, ok := g.lowerResolvedMethodDispatch(ctx, call, expected, method, objExpr, objType, callNode)
						if ok {
							lines := append(objLines, methodLines...)
							return lines, expr, retType, true
						}
					}
					if ifaceLines, expr, retType, ok := g.lowerNativeInterfaceGenericMethodDispatch(ctx, call, expected, objExpr, objType, ident.Name, callNode); ok {
						lines := append(objLines, ifaceLines...)
						return lines, expr, retType, true
					}
					if fieldCallLines, expr, retType, ok := g.compileStaticCallableFieldCall(ctx, call, expected, callee, objExpr, objType, callNode); ok {
						lines := append(objLines, fieldCallLines...)
						return lines, expr, retType, true
					}
				}
			}
		}
		if callee, ok := call.Callee.(*ast.Identifier); ok && callee != nil {
			// Generic calls on local values (e.g. generic lambdas) must call the
			// bound value, not global name lookup.
			if binding, ok := ctx.lookup(callee.Name); ok {
				binding = g.recoverDispatchBinding(ctx, binding)
				if isFunctionTypeExpr(binding.TypeExpr) {
					return g.compileFnParamCall(ctx, call, expected, binding, callNode)
				}
				if lines, expr, retType, ok := g.compileStaticApplyCall(ctx, call, expected, binding.GoName, binding.GoType, callNode); ok {
					return lines, expr, retType, true
				}
				return g.compileDynamicCall(ctx, call, expected, "", callNode)
			}
			if _, ok := g.runtimeHelperImpl(callee.Name); ok {
				return g.compileDynamicCall(ctx, call, expected, callee.Name, callNode)
			}
			if info, overload, ok := g.resolveStaticCallable(ctx, callee.Name); ok {
				if lines, expr, retType, ok := g.compileStaticNamedFunctionCall(ctx, call, expected, callee.Name, info, overload, callNode); ok {
					return lines, expr, retType, true
				}
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
		if _, ok := g.runtimeHelperImpl(callee.Name); ok {
			if binding, found := ctx.lookup(callee.Name); found {
				binding = g.recoverDispatchBinding(ctx, binding)
				if isFunctionTypeExpr(binding.TypeExpr) || g.nativeCallableInfoForGoType(binding.GoType) != nil {
					return g.compileFnParamCall(ctx, call, expected, binding, callNode)
				}
				if lines, expr, retType, ok := g.compileStaticApplyCall(ctx, call, expected, binding.GoName, binding.GoType, callNode); ok {
					return lines, expr, retType, true
				}
			}
			return g.compileDynamicCall(ctx, call, expected, callee.Name, callNode)
		}
		if info, overload, ok := g.resolveStaticCallable(ctx, callee.Name); ok {
			if lines, expr, retType, ok := g.compileStaticNamedFunctionCall(ctx, call, expected, callee.Name, info, overload, callNode); ok {
				return lines, expr, retType, true
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
		binding = g.recoverDispatchBinding(ctx, binding)
		// When calling a local binding whose type is a function (e.g. a lambda
		// parameter like `predicate: T -> bool`), use the fast call path which
		// avoids interface unwrapping, partial application checks, and overload
		// resolution overhead.
		if isFunctionTypeExpr(binding.TypeExpr) || g.nativeCallableInfoForGoType(binding.GoType) != nil {
			return g.compileFnParamCall(ctx, call, expected, binding, callNode)
		}
		if lines, expr, retType, ok := g.compileStaticApplyCall(ctx, call, expected, binding.GoName, binding.GoType, callNode); ok {
			return lines, expr, retType, true
		}
	}
	return g.compileDynamicCall(ctx, call, expected, "", callNode)
}

// compileFnParamCall compiles a call to a local binding known to be a function
// value (e.g. a lambda parameter like `predicate: T -> bool`). Uses the fast
// call path that type-switches directly on NativeFunctionValue, skipping
// interface unwrapping, partial application, thunk, and overload resolution.
func (g *generator) compileFnParamCall(ctx *compileContext, call *ast.FunctionCall, expected string, binding paramInfo, callNode string) ([]string, string, string, bool) {
	binding = g.recoverDispatchBinding(ctx, binding)
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
		argConvLines, valueExpr, ok := g.lowerRuntimeValue(ctx, expr, goType)
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
	controlLines, ok := g.lowerControlCheck(ctx, controlTemp)
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
		convLines, converted, ok := g.lowerExpectRuntimeValue(ctx, resultTemp, expected)
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
					if method, ok := g.resolveStaticMethodCallForCall(ctx, call, callee.Object, ident.Name); ok {
						method = g.concreteStaticMethodCallInfo(ctx, call, method, callee.Object, expected)
						return g.lowerResolvedMethodDispatch(ctx, call, expected, method, "", "", callNode)
					}
				}
			}
			var preferredReceiverTypeExpr ast.TypeExpression
			if callee.Member != nil && !callee.Safe {
				if ident, ok := callee.Member.(*ast.Identifier); ok && ident != nil && ident.Name != "" {
					preferredReceiverTypeExpr = g.preferredDispatchReceiverTypeExpr(ctx, call, callee.Object, ident.Name, expected)
				}
			}
			objLines, objExpr, objType, ok := g.compileDispatchReceiverExprWithExpectedTypeExpr(ctx, callee.Object, "", preferredReceiverTypeExpr)
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
					if method := g.methodForReceiver(objType, ident.Name); method != nil {
						method = g.concreteMethodCallInfo(ctx, call, method, callee.Object, objType, expected)
						methodLines, v, t, ok := g.lowerResolvedMethodDispatch(ctx, call, expected, method, objExpr, objType, callNode)
						lines = append(lines, methodLines...)
						return lines, v, t, ok
					}
					if siblingLines, expr, retType, ok := g.compileDirectImplSiblingMethodCall(ctx, call, callee, expected, objExpr, objType, callNode); ok {
						if !g.usesImplicitReceiverSiblingCall(ctx, callee.Object) {
							lines = append(lines, siblingLines...)
						} else {
							lines = append(lines[:len(lines)-len(objLines)], siblingLines...)
						}
						return lines, expr, retType, true
					}
					if method := g.compileableInterfaceMethodForConcreteReceiverExpr(ctx, callee.Object, objType, ident.Name); method != nil {
						method = g.concreteMethodCallInfo(ctx, call, method, callee.Object, objType, expected)
						methodLines, v, t, ok := g.lowerResolvedMethodDispatch(ctx, call, expected, method, objExpr, objType, callNode)
						lines = append(lines, methodLines...)
						return lines, v, t, ok
					}
					if ifaceLines, expr, retType, ok := g.compileConcreteNativeInterfaceMethodCall(ctx, call, expected, objExpr, objType, ident.Name, callNode); ok {
						lines = append(lines, ifaceLines...)
						return lines, expr, retType, true
					}
					if ifaceLines, expr, retType, ok := g.lowerNativeInterfaceMethodDispatch(ctx, call, expected, objExpr, objType, ident.Name, callNode); ok {
						lines = append(lines, ifaceLines...)
						return lines, expr, retType, true
					}
					if ifaceLines, expr, retType, ok := g.lowerNativeInterfaceGenericMethodDispatch(ctx, call, expected, objExpr, objType, ident.Name, callNode); ok {
						lines = append(lines, ifaceLines...)
						return lines, expr, retType, true
					}
					if fieldCallLines, expr, retType, ok := g.compileStaticCallableFieldCall(ctx, call, expected, callee, objExpr, objType, callNode); ok {
						lines = append(lines, fieldCallLines...)
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
					if g.hasUncompiledReceiverMethod(objType, ident.Name) {
						ctx.setReason("unsupported method call")
						return nil, "", "", false
					}
				}
			}
			if callee.Safe {
				cat := g.typeCategory(objType)
				if cat == "runtime" || cat == "any" || g.goTypeHasNilZeroValue(objType) {
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
								if g.canSpecializeImplicitReceiver(ctx) {
									method = g.concreteMethodCallInfo(ctx, call, method, callee.Object, objType, expected)
								}
								methodLines, v, t, ok := g.lowerResolvedMethodDispatch(ctx, call, expected, method, objExpr, objType, callNode)
								if ok {
									lines = append(lines, methodLines...)
									return lines, v, t, true
								}
							}
							// Fall back to dynamic dispatch
							objConvLines, objValue, ok := g.lowerRuntimeValue(ctx, objExpr, objType)
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
				objConvLines, objValue, ok := g.lowerRuntimeValue(ctx, objExpr, objType)
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
			if recoverLines, recoveredExpr, recoveredType, recovered := g.recoverDispatchExpr(ctx, call.Callee, calleeExpr, calleeType); recovered {
				lines = append(lines, recoverLines...)
				calleeExpr = recoveredExpr
				calleeType = recoveredType
			}
			if applyLines, expr, retType, ok := g.compileStaticApplyCall(ctx, call, expected, calleeExpr, calleeType, callNode); ok {
				lines = append(lines, applyLines...)
				return lines, expr, retType, true
			}
			calleeConvLines, calleeValue, ok := g.lowerRuntimeValue(ctx, calleeExpr, calleeType)
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
		argConvLines, valueExpr, ok := g.lowerRuntimeValue(ctx, expr, goType)
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
		if helper, ok := g.runtimeHelperImpl(calleeName); ok {
			var ok bool
			lines, callExpr, ok = g.appendRuntimeHelperErrorLines(ctx, lines, fmt.Sprintf("%s(%s)", helper, argList), callNode)
			if !ok {
				return nil, "", "", false
			}
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
		controlLines, ok := g.lowerControlCheck(ctx, controlTemp)
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
		baseName, ok := g.structHelperName(writebackObjType)
		if !ok {
			baseName, ok = g.structBaseName(writebackObjType)
		}
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
		controlLines, ok := g.lowerControlCheck(ctx, controlTemp)
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
			convLines, converted, ok := g.lowerExpectRuntimeValue(ctx, resultTemp, expected)
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
		convLines, converted, ok := g.lowerExpectRuntimeValue(ctx, callExpr, expected)
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

func (g *generator) dynamicMemberCallReceiverGoType(ctx *compileContext, receiver ast.Expression, receiverGoType string) string {
	if g == nil {
		return ""
	}
	if receiverGoType != "" && receiverGoType != "runtime.Value" && receiverGoType != "any" {
		return receiverGoType
	}
	if ctx != nil {
		if receiverTypeExpr, ok := g.staticReceiverTypeExpr(ctx, receiver, receiverGoType); ok && receiverTypeExpr != nil {
			if recovered, ok := g.recoverRepresentableCarrierType(ctx.packageName, receiverTypeExpr, ""); ok &&
				recovered != "" &&
				recovered != "runtime.Value" &&
				recovered != "any" {
				return recovered
			}
		}
		if receiverTypeExpr := g.dispatchReceiverTypeExpr(ctx, receiver); receiverTypeExpr != nil {
			if recovered, ok := g.recoverRepresentableCarrierType(ctx.packageName, receiverTypeExpr, ""); ok &&
				recovered != "" &&
				recovered != "runtime.Value" &&
				recovered != "any" {
				return recovered
			}
		}
	}
	return g.inferredDispatchReceiverType(ctx, receiver)
}

func (g *generator) dynamicMemberCallArgExpectations(ctx *compileContext, call *ast.FunctionCall, receiver ast.Expression, receiverGoType string, methodName string, expected string) []dynamicCallArgExpectation {
	if g == nil || ctx == nil || call == nil || receiver == nil || methodName == "" {
		return nil
	}
	receiverPkgName := ctx.packageName
	var receiverTypeExpr ast.TypeExpression
	if inferred, ok := g.staticReceiverTypeExpr(ctx, receiver, receiverGoType); ok && inferred != nil {
		receiverPkgName, receiverTypeExpr = g.typeExprContextInContext(ctx, inferred)
	} else if inferred := g.dispatchReceiverTypeExpr(ctx, receiver); inferred != nil {
		receiverPkgName, receiverTypeExpr = g.typeExprContextInContext(ctx, inferred)
	}
	receiverGoType = g.dynamicMemberCallReceiverGoType(ctx, receiver, receiverGoType)
	if receiverGoType == "" || receiverGoType == "runtime.Value" || receiverGoType == "any" {
		if receiverTypeExpr == nil {
			return nil
		}
	}
	appendMethodExpectations := func(method *methodInfo) []dynamicCallArgExpectation {
		if method == nil || method.Info == nil {
			return nil
		}
		method = g.concreteMethodCallInfo(ctx, call, method, receiver, receiverGoType, expected)
		if method == nil || method.Info == nil {
			return nil
		}
		paramOffset := 0
		if method.ExpectsSelf {
			paramOffset = 1
		}
		if paramOffset > len(method.Info.Params) {
			return nil
		}
		params := method.Info.Params[paramOffset:]
		optionalLast := g.hasOptionalLastParam(method.Info)
		if len(call.Arguments) != len(params) && !(optionalLast && len(call.Arguments) == len(params)-1) {
			return nil
		}
		expectations := make([]dynamicCallArgExpectation, 0, len(call.Arguments))
		for idx := range call.Arguments {
			if idx >= len(params) {
				return nil
			}
			param := params[idx]
			expectations = append(expectations, dynamicCallArgExpectation{
				goType:   g.staticParamCarrierType(ctx, param),
				typeExpr: param.TypeExpr,
			})
		}
		return expectations
	}
	if method := g.methodForReceiver(receiverGoType, methodName); method != nil {
		if expectations := appendMethodExpectations(method); len(expectations) > 0 {
			return expectations
		}
	}
	if receiverTypeExpr != nil {
		if baseName, ok := typeExprBaseName(receiverTypeExpr); ok && baseName != "" {
			if method := g.methodForTypeNameInPackage(receiverPkgName, baseName, methodName, true); method != nil {
				if expectations := appendMethodExpectations(method); len(expectations) > 0 {
					return expectations
				}
			}
		}
	}
	if method := g.compileableInterfaceMethodForConcreteReceiverExpr(ctx, receiver, receiverGoType, methodName); method != nil {
		if expectations := appendMethodExpectations(method); len(expectations) > 0 {
			return expectations
		}
	}
	if method, ok := g.nativeInterfaceMethodForGoType(receiverGoType, methodName); ok && method != nil {
		if len(call.Arguments) != len(method.ParamTypeExprs) {
			return nil
		}
		expectations := make([]dynamicCallArgExpectation, 0, len(call.Arguments))
		for idx := range call.Arguments {
			goType := ""
			if idx < len(method.ParamGoTypes) {
				goType = method.ParamGoTypes[idx]
			}
			typeExpr := ast.TypeExpression(nil)
			if idx < len(method.ParamTypeExprs) {
				typeExpr = method.ParamTypeExprs[idx]
			}
			expectations = append(expectations, dynamicCallArgExpectation{
				goType:   goType,
				typeExpr: typeExpr,
			})
		}
		return expectations
	}
	return nil
}

func (g *generator) compileDynamicCallArgumentRuntimeValue(ctx *compileContext, arg ast.Expression, expectation dynamicCallArgExpectation) ([]string, string, bool) {
	if g == nil || ctx == nil || arg == nil {
		return nil, "", false
	}
	expectedGoType := expectation.goType
	compileExpectedGoType := expectedGoType
	if g.nativeUnionInfoForGoType(expectedGoType) != nil {
		compileExpectedGoType = ""
	}
	if ifaceType, ok := g.interfaceTypeExpr(expectation.typeExpr); ok &&
		(compileExpectedGoType == "" || compileExpectedGoType == "runtime.Value" || compileExpectedGoType == "any") {
		if ifaceInfo, ok := g.ensureNativeInterfaceInfo(ctx.packageName, ifaceType); ok && ifaceInfo != nil && ifaceInfo.GoType != "" {
			expectedGoType = ifaceInfo.GoType
			compileExpectedGoType = ifaceInfo.GoType
		}
	}
	var (
		lines  []string
		expr   string
		goType string
		ok     bool
	)
	if compileExpectedGoType != "" || expectation.typeExpr != nil {
		lines, expr, goType, ok = g.compileExprLinesWithExpectedTypeExpr(ctx, arg, compileExpectedGoType, expectation.typeExpr)
	} else {
		previousExpectedTypeExpr := ctx.expectedTypeExpr
		ctx.expectedTypeExpr = nil
		lines, expr, goType, ok = g.compileExprLines(ctx, arg, "")
		ctx.expectedTypeExpr = previousExpectedTypeExpr
	}
	if !ok {
		return nil, "", false
	}
	argConvLines, valueExpr, ok := g.lowerRuntimeValue(ctx, expr, goType)
	if !ok {
		ctx.setReason("call argument unsupported")
		return nil, "", false
	}
	lines = append(lines, argConvLines...)
	temp := ctx.newTemp()
	lines = append(lines, fmt.Sprintf("%s := %s", temp, valueExpr))
	return lines, temp, true
}

func (g *generator) compileStaticCallableFieldCall(ctx *compileContext, call *ast.FunctionCall, expected string, callee *ast.MemberAccessExpression, objExpr string, objType string, callNode string) ([]string, string, string, bool) {
	if g == nil || ctx == nil || call == nil || callee == nil || callee.Safe || objExpr == "" || objType == "" {
		return nil, "", "", false
	}
	if fieldLines, fieldExpr, fieldType, ok := g.compileOriginStructFieldAccess(ctx, callee, objExpr, ""); ok {
		if applyLines, expr, retType, ok := g.compileStaticApplyCall(ctx, call, expected, fieldExpr, fieldType, callNode); ok {
			fieldLines = append(fieldLines, applyLines...)
			return fieldLines, expr, retType, true
		}
	}
	info := g.structInfoByGoName(objType)
	if info == nil {
		return nil, "", "", false
	}
	field, ok := g.structFieldForMember(info, callee.Member)
	if !ok || field == nil {
		return nil, "", "", false
	}
	return g.compileStaticApplyCall(ctx, call, expected, fmt.Sprintf("%s.%s", objExpr, field.GoName), field.GoType, callNode)
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
	receiverExpr := objExpr
	receiverType := objType
	if ctx.implicitReceiver.GoName != "" && ctx.implicitReceiver.GoType != "" {
		receiverExpr = ctx.implicitReceiver.GoName
		receiverType = ctx.implicitReceiver.GoType
	}
	// Prefer full concrete receiver resolution over the cached sibling entry.
	// Default/generic sibling metadata can still point at a base impl, while the
	// implicit receiver may already be fully specialized on the current body.
	preferConcreteSibling := g.canSpecializeImplicitReceiver(ctx) && !g.typeExprHasGeneric(ctx.returnTypeExpr, ctx.genericNames)
	if preferConcreteSibling && receiverType != "" && receiverType != "runtime.Value" && receiverType != "any" {
		if method := g.methodForReceiver(receiverType, ident.Name); method != nil {
			method = g.concreteMethodCallInfo(ctx, call, method, callee.Object, receiverType, expected)
			if lines, expr, retType, ok := g.lowerResolvedMethodDispatch(ctx, call, expected, method, receiverExpr, receiverType, callNode); ok {
				return lines, expr, retType, true
			}
		}
		if method := g.compileableInterfaceMethodForConcreteReceiverExpr(ctx, callee.Object, receiverType, ident.Name); method != nil {
			method = g.concreteMethodCallInfo(ctx, call, method, callee.Object, receiverType, expected)
			if lines, expr, retType, ok := g.lowerResolvedMethodDispatch(ctx, call, expected, method, receiverExpr, receiverType, callNode); ok {
				return lines, expr, retType, true
			}
		}
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
	if g.nativeInterfaceDefaultByInfo[sibling.Info] == nil && g.canSpecializeImplicitReceiver(ctx) {
		method = g.concreteMethodCallInfo(ctx, call, method, callee.Object, objType, expected)
	}
	return g.lowerResolvedMethodDispatch(ctx, call, expected, method, receiverExpr, receiverType, callNode)
}

func (g *generator) usesImplicitReceiverSiblingCall(ctx *compileContext, object ast.Expression) bool {
	if ctx == nil || !ctx.hasImplicitReceiver {
		return false
	}
	ident, ok := object.(*ast.Identifier)
	if !ok || ident == nil {
		return false
	}
	return ident.Name != "" && ident.Name == ctx.implicitReceiver.Name
}

func (g *generator) canSpecializeImplicitReceiver(ctx *compileContext) bool {
	if g == nil || ctx == nil || !ctx.hasImplicitReceiver || ctx.implicitReceiver.TypeExpr == nil {
		return false
	}
	receiverTypeExpr := normalizeTypeExprForPackage(g, ctx.packageName, ctx.implicitReceiver.TypeExpr)
	if receiverTypeExpr == nil || !g.typeExprFullyBound(ctx.packageName, receiverTypeExpr) {
		return false
	}
	if ctx.implicitReceiver.GoType == "" {
		return true
	}
	canonicalGoType, ok := g.lowerCarrierTypeInPackage(ctx.packageName, receiverTypeExpr)
	if !ok || canonicalGoType == "" || canonicalGoType == "runtime.Value" || canonicalGoType == "any" {
		return false
	}
	return canonicalGoType == ctx.implicitReceiver.GoType
}

func (g *generator) runtimeHelperImpl(name string) (string, bool) {
	switch name {
	case "__able_array_new":
		return "__able_array_new_impl", true
	case "__able_array_with_capacity":
		return "__able_array_with_capacity_impl", true
	case "__able_array_size":
		return "__able_array_size_impl", true
	case "__able_array_capacity":
		return "__able_array_capacity_impl", true
	case "__able_array_set_len":
		return "__able_array_set_len_impl", true
	case "__able_array_read":
		return "__able_array_read_impl", true
	case "__able_array_write":
		return "__able_array_write_impl", true
	case "__able_array_reserve":
		return "__able_array_reserve_impl", true
	case "__able_array_clone":
		return "__able_array_clone_impl", true
	case "__able_hash_map_new":
		return "__able_hash_map_new_impl", true
	case "__able_hash_map_with_capacity":
		return "__able_hash_map_with_capacity_impl", true
	case "__able_hash_map_get":
		return "__able_hash_map_get_impl", true
	case "__able_hash_map_set":
		return "__able_hash_map_set_impl", true
	case "__able_hash_map_remove":
		return "__able_hash_map_remove_impl", true
	case "__able_hash_map_contains":
		return "__able_hash_map_contains_impl", true
	case "__able_hash_map_size":
		return "__able_hash_map_size_impl", true
	case "__able_hash_map_clear":
		return "__able_hash_map_clear_impl", true
	case "__able_hash_map_for_each":
		return "__able_hash_map_for_each_impl", true
	case "__able_hash_map_clone":
		return "__able_hash_map_clone_impl", true
	case "__able_String_from_builtin":
		return "__able_string_from_builtin_impl", true
	case "__able_String_to_builtin":
		return "__able_string_to_builtin_impl", true
	case "__able_char_from_codepoint":
		return "__able_char_from_codepoint_impl", true
	case "__able_char_to_codepoint":
		return "__able_char_to_codepoint_impl", true
	case "__able_ratio_from_float":
		return "__able_ratio_from_float_impl", true
	case "__able_f32_bits":
		return "__able_f32_bits_impl", true
	case "__able_f64_bits":
		return "__able_f64_bits_impl", true
	case "__able_u64_mul":
		return "__able_u64_mul_impl", true
	case "__able_channel_new":
		return "__able_channel_new_impl", true
	case "__able_channel_send":
		return "__able_channel_send_impl", true
	case "__able_channel_receive":
		return "__able_channel_receive_impl", true
	case "__able_channel_try_send":
		return "__able_channel_try_send_impl", true
	case "__able_channel_try_receive":
		return "__able_channel_try_receive_impl", true
	case "__able_channel_await_try_recv":
		return "__able_channel_await_try_recv_impl", true
	case "__able_channel_await_try_send":
		return "__able_channel_await_try_send_impl", true
	case "__able_channel_close":
		return "__able_channel_close_impl", true
	case "__able_channel_is_closed":
		return "__able_channel_is_closed_impl", true
	case "__able_mutex_new":
		return "__able_mutex_new_impl", true
	case "__able_mutex_lock":
		return "__able_mutex_lock_impl", true
	case "__able_mutex_unlock":
		return "__able_mutex_unlock_impl", true
	case "__able_mutex_await_lock":
		return "__able_mutex_await_lock_impl", true
	default:
		return "", false
	}
}

func (g *generator) compileSafeMemberCall(ctx *compileContext, call *ast.FunctionCall, callee *ast.MemberAccessExpression, expected string, objExpr string, objType string, callNode string) ([]string, string, string, bool) {
	if call == nil || callee == nil {
		ctx.setReason("missing safe member call")
		return nil, "", "", false
	}
	if objType != "runtime.Value" && objType != "any" {
		if lines, expr, resultType, ok := g.compileStaticSafeMemberCall(ctx, call, callee, expected, objExpr, objType, callNode); ok {
			return lines, expr, resultType, true
		}
	}
	objConvLines, objValue, ok := g.lowerRuntimeValue(ctx, objExpr, objType)
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
		convLines, valueExpr, ok := g.lowerRuntimeValue(ctx, expr, goType)
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
	} else if expected != "" {
		resultType = expected
	}
	resultTemp := ctx.newTemp()
	objTemp := ctx.newTemp()
	nilExpr := safeNilReturnExpr(resultType)
	if wrapped, ok := g.nativeUnionNilExpr(resultType); ok {
		nilExpr = wrapped
	}
	lines := make([]string, 0, len(argPreLines)+len(call.Arguments)*2+8+len(objConvLines))
	lines = append(lines, objConvLines...)
	lines = append(lines, fmt.Sprintf("%s := %s", objTemp, objValue))
	lines = append(lines, fmt.Sprintf("var %s %s", resultTemp, resultType))
	lines = append(lines, fmt.Sprintf("if %s {", g.safeNavigationNilCheckExpr(objTemp, objType)))
	lines = append(lines, fmt.Sprintf("%s = %s", resultTemp, nilExpr))
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
		if resultType == "runtime.Value" {
			lines = append(lines, fmt.Sprintf("%s = %s", resultTemp, callExpr))
		} else {
			convLines, converted, ok := g.safeNavigationCoerceSuccessExpr(ctx, callExpr, "runtime.Value", resultType)
			if !ok {
				ctx.setReason("call return type mismatch")
				return nil, "", "", false
			}
			lines = append(lines, convLines...)
			lines = append(lines, fmt.Sprintf("%s = %s", resultTemp, converted))
		}
	}
	lines = append(lines, "}")
	if expected != "" && expected != "runtime.Value" && !g.isVoidType(expected) && resultType == "runtime.Value" {
		convLines, converted, ok := g.lowerExpectRuntimeValue(ctx, resultTemp, expected)
		if !ok {
			ctx.setReason("call return type mismatch")
			return nil, "", "", false
		}
		lines = append(lines, convLines...)
		return lines, converted, expected, true
	}
	return lines, resultTemp, resultType, true
}
