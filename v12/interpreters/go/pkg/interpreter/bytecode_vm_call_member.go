package interpreter

import (
	"fmt"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func bytecodeCanDirectMemberCall(receiver runtime.Value, memberName string) bool {
	if memberName == "" {
		return false
	}
	switch v := receiver.(type) {
	case runtime.StringValue:
		return true
	case *runtime.StringValue:
		return v != nil
	case *runtime.ArrayValue:
		return v != nil
	case *runtime.IteratorValue:
		return v != nil
	case *runtime.FutureValue:
		return v != nil
	case *runtime.HasherValue:
		return v != nil
	case runtime.ErrorValue:
		return true
	case *runtime.ErrorValue:
		return v != nil
	case *runtime.StructInstanceValue:
		if v == nil {
			return false
		}
		if val, ok := v.Fields[memberName]; ok && isCallableRuntimeValue(val) {
			return false
		}
		return true
	default:
		return false
	}
}

func bytecodeResolveExactInjectedNativeCallTarget(callable runtime.Value, receiver runtime.Value, explicitArgCount int) (bytecodeExactNativeCallTarget, bool) {
	switch fn := callable.(type) {
	case runtime.NativeFunctionValue:
		if fn.Arity >= 0 && explicitArgCount != fn.Arity {
			return bytecodeExactNativeCallTarget{}, false
		}
		return bytecodeExactNativeCallTarget{
			native:           fn,
			injectedReceiver: receiver,
			hasReceiver:      true,
		}, true
	case *runtime.NativeFunctionValue:
		if fn == nil {
			return bytecodeExactNativeCallTarget{}, false
		}
		if fn.Arity >= 0 && explicitArgCount != fn.Arity {
			return bytecodeExactNativeCallTarget{}, false
		}
		return bytecodeExactNativeCallTarget{
			native:           *fn,
			injectedReceiver: receiver,
			hasReceiver:      true,
		}, true
	case runtime.NativeBoundMethodValue:
		if fn.Method.Arity >= 0 && explicitArgCount != fn.Method.Arity {
			return bytecodeExactNativeCallTarget{}, false
		}
		return bytecodeExactNativeCallTarget{
			native:           fn.Method,
			injectedReceiver: fn.Receiver,
			hasReceiver:      true,
		}, true
	case *runtime.NativeBoundMethodValue:
		if fn == nil {
			return bytecodeExactNativeCallTarget{}, false
		}
		if fn.Method.Arity >= 0 && explicitArgCount != fn.Method.Arity {
			return bytecodeExactNativeCallTarget{}, false
		}
		return bytecodeExactNativeCallTarget{
			native:           fn.Method,
			injectedReceiver: fn.Receiver,
			hasReceiver:      true,
		}, true
	default:
		return bytecodeExactNativeCallTarget{}, false
	}
}

func (vm *bytecodeVM) callResolvedCallableWithInjectedReceiver(callable runtime.Value, receiver runtime.Value, explicitArgs []runtime.Value, callNode *ast.FunctionCall) (runtime.Value, error) {
	if vm == nil || vm.interp == nil {
		return nil, fmt.Errorf("bytecode VM is nil")
	}
	return vm.interp.callCallableValueWithInjectedReceiver(callable, receiver, explicitArgs, vm.env, callNode, true)
}

func (vm *bytecodeVM) resolveConcreteMemberOverload(callable runtime.Value, receiver runtime.Value, explicitArgs []runtime.Value, callNode *ast.FunctionCall) (*runtime.FunctionValue, runtime.Value, bool, error) {
	if vm == nil || vm.interp == nil {
		return nil, nil, false, nil
	}

	var (
		overloads        []*runtime.FunctionValue
		injectedReceiver runtime.Value
		hasReceiver      bool
	)
	switch fn := callable.(type) {
	case *runtime.FunctionOverloadValue:
		overloads = functionOverloadsView(fn)
		injectedReceiver = receiver
		hasReceiver = true
	case runtime.BoundMethodValue:
		overloads = functionOverloadsView(fn.Method)
		injectedReceiver = fn.Receiver
		hasReceiver = true
	case *runtime.BoundMethodValue:
		if fn == nil {
			return nil, nil, false, fmt.Errorf("bound method is nil")
		}
		overloads = functionOverloadsView(fn.Method)
		injectedReceiver = fn.Receiver
		hasReceiver = true
	default:
		return nil, nil, false, nil
	}

	if len(overloads) == 0 {
		return nil, nil, false, nil
	}

	selected := overloads[0]
	if len(overloads) > 1 {
		evalArgs := explicitArgs
		if hasReceiver {
			totalArgs := len(explicitArgs) + 1
			var inline [overloadArgSignatureInlineLimit + 1]runtime.Value
			if totalArgs <= len(inline) {
				inline[0] = injectedReceiver
				copy(inline[1:totalArgs], explicitArgs)
				evalArgs = inline[:totalArgs]
			} else {
				evalArgs = make([]runtime.Value, totalArgs)
				evalArgs[0] = injectedReceiver
				copy(evalArgs[1:], explicitArgs)
			}
		}
		if len(evalArgs) < minArgsForOverloads(overloads) {
			return nil, nil, false, nil
		}
		var err error
		selected, err = vm.interp.selectRuntimeOverload(overloads, evalArgs, callNode)
		if err != nil {
			return nil, nil, false, err
		}
		if selected == nil {
			return nil, nil, false, nil
		}
	}

	return selected, injectedReceiver, hasReceiver, nil
}

func (vm *bytecodeVM) tryInlineMemberCallableFromStack(callable runtime.Value, receiver runtime.Value, argBase int, argCount int, truncateTo int, callNode *ast.FunctionCall, currentProgram *bytecodeProgram) (*bytecodeProgram, error) {
	if fn, injectedReceiver, hasInjectedReceiver, ok := inlineCallFunctionValue(callable); ok {
		if !hasInjectedReceiver {
			injectedReceiver = receiver
			hasInjectedReceiver = true
		}
		return vm.tryInlineResolvedCallFromStack(fn, injectedReceiver, hasInjectedReceiver, argBase, argCount, truncateTo, callNode, currentProgram)
	}
	return vm.tryInlineCallFromStack(callable, argBase, argCount, truncateTo, callNode, currentProgram)
}

func (vm *bytecodeVM) execCachedResolvedMemberCall(callee runtime.Value, memberName string, receiverIndex int, argBase int, argCount int, callNode *ast.FunctionCall, currentProgram *bytecodeProgram) (*bytecodeProgram, error) {
	var traceNode ast.Node
	if callNode != nil {
		traceNode = callNode
	}
	receiver := vm.stack[receiverIndex]
	explicitArgs := vm.stack[argBase:]
	if overloadFn, overloadReceiver, ok, err := vm.resolveConcreteMemberOverload(callee, receiver, explicitArgs, callNode); err != nil {
		return nil, err
	} else if ok {
		if newProg, err := vm.tryInlineResolvedCallFromStack(overloadFn, overloadReceiver, true, argBase, argCount, receiverIndex, callNode, currentProgram); err != nil {
			return nil, err
		} else if newProg != nil {
			if vm.interp != nil {
				vm.interp.recordBytecodeCallTrace("call_member", memberName, "resolved_method", "inline", traceNode)
			}
			if vm.interp != nil && vm.interp.bytecodeStatsEnabled {
				vm.interp.recordBytecodeInlineCallHit()
			}
			return newProg, nil
		} else if vm.interp != nil && vm.interp.bytecodeStatsEnabled {
			vm.interp.recordBytecodeInlineCallMiss()
		}
		vm.stack = vm.stack[:receiverIndex]
		result, err := vm.callResolvedCallableWithInjectedReceiver(overloadFn, overloadReceiver, explicitArgs, callNode)
		if vm.interp != nil {
			vm.interp.recordBytecodeCallTrace("call_member", memberName, "resolved_method", "generic", traceNode)
		}
		return vm.finishCompletedCall(result, err, callNode, nil)
	}
	if target, ok := bytecodeResolveExactNativeCallTarget(callee, argCount); ok {
		if vm.interp != nil {
			vm.interp.recordBytecodeCallTrace("call_member", memberName, "resolved_method", "exact_native", traceNode)
		}
		vm.stack = vm.stack[:receiverIndex]
		result, _, err := vm.execExactNativeCall(target, explicitArgs, callNode)
		return vm.finishCompletedCall(result, err, callNode, nil)
	}
	if newProg, err := vm.tryInlineMemberCallableFromStack(callee, receiver, argBase, argCount, receiverIndex, callNode, currentProgram); err != nil {
		return nil, err
	} else if newProg != nil {
		if vm.interp != nil {
			vm.interp.recordBytecodeCallTrace("call_member", memberName, "resolved_method", "inline", traceNode)
		}
		if vm.interp != nil && vm.interp.bytecodeStatsEnabled {
			vm.interp.recordBytecodeInlineCallHit()
		}
		return newProg, nil
	} else if vm.interp != nil && vm.interp.bytecodeStatsEnabled {
		vm.interp.recordBytecodeInlineCallMiss()
	}
	vm.stack = vm.stack[:receiverIndex]
	if bytecodeCallTargetNeedsStableArgs(callee) {
		explicitArgs = copyCallArgs(explicitArgs)
	}
	if vm.interp != nil {
		vm.interp.recordBytecodeCallTrace("call_member", memberName, "resolved_method", "generic", traceNode)
	}
	result, err := vm.interp.callCallableValueMutable(callee, explicitArgs, vm.env, callNode)
	return vm.finishCompletedCall(result, err, callNode, nil)
}

// execCallMember handles bytecodeOpCallMember for the common `obj.method(...)`
// syntax path without materializing an intermediate bound-method value.
func (vm *bytecodeVM) execCallMember(instr bytecodeInstruction, currentProgram *bytecodeProgram) (*bytecodeProgram, error) {
	if instr.argCount < 0 {
		return nil, fmt.Errorf("bytecode call-member arg count invalid")
	}
	if len(vm.stack) < instr.argCount+1 {
		return nil, fmt.Errorf("bytecode stack underflow")
	}
	if instr.name == "" {
		return nil, fmt.Errorf("bytecode call-member missing member name")
	}
	receiverIndex := len(vm.stack) - instr.argCount - 1
	argBase := receiverIndex + 1
	receiver := vm.stack[receiverIndex]
	if instr.safe && isNilRuntimeValue(receiver) {
		vm.stack = vm.stack[:receiverIndex]
		vm.stack = append(vm.stack, runtime.NilValue{})
		vm.ip++
		return nil, nil
	}
	var callNode *ast.FunctionCall
	if instr.node != nil {
		if call, ok := instr.node.(*ast.FunctionCall); ok {
			callNode = call
		}
	}
	traceNode := instr.node
	if callNode != nil {
		traceNode = callNode
	}
	statsEnabled := vm.interp != nil && vm.interp.bytecodeStatsEnabled
	useMethodCache := vm.canUseMemberMethodCache(instr.name, true)

	if bytecodeCanDirectMemberCall(receiver, instr.name) {
		if useMethodCache {
			if cached, ok := vm.lookupCachedMemberMethodEntry(currentProgram, vm.ip, instr.name, true, receiver); ok {
				if newProg, handled, err := vm.execCachedMemberMethodFastPath(cached.fastPath, instr, receiverIndex, argBase, callNode); handled {
					return newProg, err
				}
				if callable, ok := cached.boundCallable(receiver); ok {
					return vm.execCachedResolvedMemberCall(callable, instr.name, receiverIndex, argBase, instr.argCount, callNode, currentProgram)
				}
			}
		}
		callable, found, err := vm.interp.resolveMethodCallableFromPool(vm.env, instr.name, receiver, "")
		if err != nil {
			return nil, vm.attachBytecodeRuntimeContext(err, callNode, nil)
		}
		if found {
			if useMethodCache {
				if bound, ok := bindMemberMethodTemplate(receiver, callable); ok {
					vm.storeCachedMemberMethod(currentProgram, vm.ip, instr.name, true, receiver, bound)
				}
			}
			if fn, ok := bytecodeResolvedMemberFastPathFunction(callable); ok {
				kind := vm.resolvedMemberMethodFastPath(instr.name, receiver, fn)
				if newProg, handled, err := vm.execCachedMemberMethodFastPath(kind, instr, receiverIndex, argBase, callNode); handled {
					return newProg, err
				}
			}
			if newProg, handled, err := vm.execCanonicalArrayGetOverloadMemberFast(callable, instr, receiverIndex, argBase, callNode); handled {
				return newProg, err
			}
			if overloadFn, overloadReceiver, ok, err := vm.resolveConcreteMemberOverload(callable, receiver, vm.stack[argBase:], callNode); err != nil {
				return nil, err
			} else if ok {
				kind := vm.resolvedMemberMethodFastPath(instr.name, overloadReceiver, overloadFn)
				if newProg, handled, err := vm.execCachedMemberMethodFastPath(kind, instr, receiverIndex, argBase, callNode); handled {
					return newProg, err
				}
				if newProg, err := vm.tryInlineResolvedCallFromStack(overloadFn, overloadReceiver, true, argBase, instr.argCount, receiverIndex, callNode, currentProgram); err != nil {
					return nil, err
				} else if newProg != nil {
					if vm.interp != nil {
						vm.interp.recordBytecodeCallTrace("call_member", instr.name, "resolved_method", "inline", traceNode)
					}
					if statsEnabled {
						vm.interp.recordBytecodeInlineCallHit()
					}
					return newProg, nil
				} else if statsEnabled {
					vm.interp.recordBytecodeInlineCallMiss()
				}
				args := vm.stack[argBase:]
				vm.stack = vm.stack[:receiverIndex]
				if vm.interp != nil {
					vm.interp.recordBytecodeCallTrace("call_member", instr.name, "resolved_method", "generic", traceNode)
				}
				result, err := vm.callResolvedCallableWithInjectedReceiver(overloadFn, overloadReceiver, args, callNode)
				return vm.finishCompletedCall(result, err, callNode, nil)
			}
			if target, ok := bytecodeResolveExactInjectedNativeCallTarget(callable, receiver, instr.argCount); ok {
				if vm.interp != nil {
					vm.interp.recordBytecodeCallTrace("call_member", instr.name, "resolved_method", "exact_native", traceNode)
				}
				args := vm.stack[argBase:]
				vm.stack = vm.stack[:receiverIndex]
				result, _, err := vm.execExactNativeCall(target, args, callNode)
				return vm.finishCompletedCall(result, err, callNode, nil)
			}
			if newProg, err := vm.tryInlineMemberCallableFromStack(callable, receiver, argBase, instr.argCount, receiverIndex, callNode, currentProgram); err != nil {
				return nil, err
			} else if newProg != nil {
				if vm.interp != nil {
					vm.interp.recordBytecodeCallTrace("call_member", instr.name, "resolved_method", "inline", traceNode)
				}
				if statsEnabled {
					vm.interp.recordBytecodeInlineCallHit()
				}
				return newProg, nil
			} else if statsEnabled {
				vm.interp.recordBytecodeInlineCallMiss()
			}
			args := vm.stack[argBase:]
			vm.stack = vm.stack[:receiverIndex]
			if vm.interp != nil {
				vm.interp.recordBytecodeCallTrace("call_member", instr.name, "resolved_method", "generic", traceNode)
			}
			result, err := vm.callResolvedCallableWithInjectedReceiver(callable, receiver, args, callNode)
			return vm.finishCompletedCall(result, err, callNode, nil)
		}
	}

	if newProg, handled, err := vm.execCanonicalStringByteIteratorNextCallMemberFast(instr, receiverIndex, callNode); handled {
		return newProg, err
	}

	memberExpr := ast.Expression(ast.ID(instr.name))
	callee, err := vm.interp.memberAccessOnValueWithOptions(receiver, memberExpr, vm.env, true)
	if err != nil {
		return nil, vm.attachBytecodeRuntimeContext(err, callNode, nil)
	}
	if instr.name == "new" && instr.argCount == 0 {
		if newProg, handled, err := vm.execStaticArrayNewMemberFast(instr, receiver, callee, receiverIndex, callNode); handled {
			return newProg, err
		}
	}
	if overloadFn, overloadReceiver, ok, err := vm.resolveConcreteMemberOverload(callee, receiver, vm.stack[argBase:], callNode); err != nil {
		return nil, err
	} else if ok {
		kind := vm.resolvedMemberMethodFastPath(instr.name, overloadReceiver, overloadFn)
		if newProg, handled, err := vm.execCachedMemberMethodFastPath(kind, instr, receiverIndex, argBase, callNode); handled {
			return newProg, err
		}
		if newProg, err := vm.tryInlineResolvedCallFromStack(overloadFn, overloadReceiver, true, argBase, instr.argCount, receiverIndex, callNode, currentProgram); err != nil {
			return nil, err
		} else if newProg != nil {
			if vm.interp != nil {
				vm.interp.recordBytecodeCallTrace("call_member", instr.name, "member_access", "inline", traceNode)
			}
			if statsEnabled {
				vm.interp.recordBytecodeInlineCallHit()
			}
			return newProg, nil
		} else if statsEnabled {
			vm.interp.recordBytecodeInlineCallMiss()
		}
		args := vm.stack[argBase:]
		vm.stack = vm.stack[:receiverIndex]
		if vm.interp != nil {
			vm.interp.recordBytecodeCallTrace("call_member", instr.name, "member_access", "generic", traceNode)
		}
		result, err := vm.callResolvedCallableWithInjectedReceiver(overloadFn, overloadReceiver, args, callNode)
		return vm.finishCompletedCall(result, err, callNode, nil)
	}
	if target, ok := bytecodeResolveExactNativeCallTarget(callee, instr.argCount); ok {
		if vm.interp != nil {
			vm.interp.recordBytecodeCallTrace("call_member", instr.name, "member_access", "exact_native", traceNode)
		}
		args := vm.stack[argBase:]
		vm.stack = vm.stack[:receiverIndex]
		result, _, err := vm.execExactNativeCall(target, args, callNode)
		return vm.finishCompletedCall(result, err, callNode, nil)
	}
	if newProg, err := vm.tryInlineMemberCallableFromStack(callee, receiver, argBase, instr.argCount, receiverIndex, callNode, currentProgram); err != nil {
		return nil, err
	} else if newProg != nil {
		if vm.interp != nil {
			vm.interp.recordBytecodeCallTrace("call_member", instr.name, "member_access", "inline", traceNode)
		}
		if statsEnabled {
			vm.interp.recordBytecodeInlineCallHit()
		}
		return newProg, nil
	} else if statsEnabled {
		vm.interp.recordBytecodeInlineCallMiss()
	}
	args := vm.stack[argBase:]
	vm.stack = vm.stack[:receiverIndex]
	if bytecodeCallTargetNeedsStableArgs(callee) {
		args = copyCallArgs(args)
	}
	if vm.interp != nil {
		vm.interp.recordBytecodeCallTrace("call_member", instr.name, "member_access", "generic", traceNode)
	}
	result, err := vm.interp.callCallableValueMutable(callee, args, vm.env, callNode)
	return vm.finishCompletedCall(result, err, callNode, nil)
}
