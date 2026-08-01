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
	case runtime.BoolValue:
		return true
	case *runtime.BoolValue:
		return v != nil
	case runtime.CharValue:
		return true
	case *runtime.CharValue:
		return v != nil
	case runtime.NilValue:
		return true
	case *runtime.NilValue:
		return v != nil
	case runtime.IntegerValue:
		return true
	case *runtime.IntegerValue:
		return v != nil
	case runtime.FloatValue:
		return true
	case *runtime.FloatValue:
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
		if val, ok := structNamedFieldValue(v, memberName); ok && isCallableRuntimeValue(val) {
			return false
		}
		return true
	case *runtime.InterfaceValue:
		return v != nil
	default:
		return false
	}
}

func (vm *bytecodeVM) resolveConcreteMemberOverload(callable runtime.Value, receiver runtime.Value, explicitArgs []runtime.Value, callNode *ast.FunctionCall) (*runtime.FunctionValue, runtime.Value, bool, error) {
	if vm == nil || vm.interp == nil {
		return nil, nil, false, nil
	}
	if fn, injectedReceiver, hasInjectedReceiver, ok := inlineCallFunctionValue(callable); ok && hasInjectedReceiver {
		return fn, injectedReceiver, true, nil
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
				inline[0] = vm.materializePrimitiveValue(bytecodeMaterializationRequiredDynamic, bytecodeMaterializationReasonInterfaceUnion, injectedReceiver)
				vm.copyMaterializedCallArgs(inline[1:totalArgs], explicitArgs, bytecodeMaterializationRequiredDynamic, bytecodeMaterializationReasonInterfaceUnion)
				evalArgs = inline[:totalArgs]
			} else {
				evalArgs = make([]runtime.Value, totalArgs)
				evalArgs[0] = vm.materializePrimitiveValue(bytecodeMaterializationRequiredDynamic, bytecodeMaterializationReasonInterfaceUnion, injectedReceiver)
				vm.copyMaterializedCallArgs(evalArgs[1:], explicitArgs, bytecodeMaterializationRequiredDynamic, bytecodeMaterializationReasonInterfaceUnion)
			}
		} else if len(explicitArgs) > 0 {
			if len(explicitArgs) <= overloadArgSignatureInlineLimit {
				var inline [overloadArgSignatureInlineLimit]runtime.Value
				vm.copyMaterializedCallArgs(inline[:len(explicitArgs)], explicitArgs, bytecodeMaterializationRequiredDynamic, bytecodeMaterializationReasonInterfaceUnion)
				evalArgs = inline[:len(explicitArgs)]
			} else {
				evalArgs = make([]runtime.Value, len(explicitArgs))
				vm.copyMaterializedCallArgs(evalArgs, explicitArgs, bytecodeMaterializationRequiredDynamic, bytecodeMaterializationReasonInterfaceUnion)
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

func (vm *bytecodeVM) execCachedResolvedMemberCall(cached bytecodeCachedMemberMethod, memberName string, receiverIndex int, argBase int, argCount int, callNode *ast.FunctionCall, currentProgram *bytecodeProgram) (*bytecodeProgram, error) {
	var traceNode ast.Node
	if callNode != nil {
		traceNode = callNode
	}
	statsEnabled := vm.interp != nil && vm.interp.bytecodeStatsEnabled
	receiver := vm.stackValue(receiverIndex)
	explicitArgs := vm.stackValuesFrom(argBase)
	callee := cached.template
	if callee == nil {
		return nil, fmt.Errorf("bytecode cached member template missing")
	}
	methodReceiver := receiver
	hasMethodReceiver := false
	if iface, ok := receiver.(*runtime.InterfaceValue); ok && iface != nil {
		methodReceiver = interfaceMethodReceiver(vm.interp, iface, callee)
		hasMethodReceiver = true
	}
	distinctInjectedReceiver := bytecodeDirectMemberCallHasDistinctInjectedReceiver(receiver, methodReceiver, hasMethodReceiver)
	if cached.dispatch == bytecodeMemberMethodDispatchExactNative {
		if target, ok := bytecodeResolveExactInjectedNativeCallTarget(callee, methodReceiver, argCount); ok {
			if vm.interp != nil {
				vm.interp.recordBytecodeCallTrace("call_member", memberName, "resolved_method", "exact_native", traceNode)
			}
			if statsEnabled {
				vm.interp.recordBytecodeCallMemberDispatch(bytecodeCallMemberStatsResolvedExactNative)
			}
			if target.native.RawImpl == nil {
				explicitArgs = vm.prepareMaterializedCallArgs(explicitArgs, false, bytecodeMaterializationRequiredDynamic, bytecodeMaterializationReasonInterfaceUnion)
			}
			vm.truncateStack(receiverIndex)
			return vm.execAndFinishExactNativeCall(target, explicitArgs, callNode)
		}
	}
	if cached.dispatch == bytecodeMemberMethodDispatchInline && cached.inlineFn != nil {
		if newProg, err := vm.tryInlineResolvedCallFromStack(cached.inlineFn, methodReceiver, true, argBase, argCount, receiverIndex, callNode, currentProgram); err != nil {
			return nil, err
		} else if newProg != nil {
			if vm.interp != nil {
				vm.interp.recordBytecodeCallTrace("call_member", memberName, "resolved_method", "inline", traceNode)
			}
			if vm.interp != nil && vm.interp.bytecodeStatsEnabled {
				vm.interp.recordBytecodeInlineCallHit()
			}
			if statsEnabled {
				vm.interp.recordBytecodeCallMemberDispatch(bytecodeCallMemberStatsResolvedInline)
			}
			return newProg, nil
		} else if vm.interp != nil && vm.interp.bytecodeStatsEnabled {
			vm.interp.recordBytecodeInlineCallMiss()
		}
		if vm.interp != nil {
			vm.interp.recordBytecodeCallTrace("call_member", memberName, "resolved_method", "generic", traceNode)
		}
		if !distinctInjectedReceiver {
			if result, handled, err := vm.tryCallResolvedCallableFromMemberStack(cached.template, methodReceiver, receiverIndex, argBase, argCount, callNode); handled || err != nil {
				if statsEnabled {
					vm.interp.recordBytecodeCallMemberDispatch(bytecodeCallMemberStatsResolvedGeneric)
				}
				return vm.finishCompletedCall(result, err, callNode, nil)
			}
		}
	}
	if !distinctInjectedReceiver {
		if newProg, handled, err := vm.execCanonicalArrayGetOverloadMemberFast(
			callee,
			bytecodeInstruction{name: memberName, argCount: argCount, node: traceNode},
			receiverIndex,
			argBase,
			callNode,
		); handled {
			return newProg, err
		}
	}
	if overloadFn, overloadReceiver, ok, err := vm.resolveConcreteMemberOverload(callee, methodReceiver, explicitArgs, callNode); err != nil {
		return nil, err
	} else if ok {
		overloadDistinctReceiver := bytecodeDirectMemberCallHasDistinctInjectedReceiver(receiver, overloadReceiver, true)
		if newProg, err := vm.tryInlineResolvedCallFromStack(overloadFn, overloadReceiver, true, argBase, argCount, receiverIndex, callNode, currentProgram); err != nil {
			return nil, err
		} else if newProg != nil {
			if vm.interp != nil {
				vm.interp.recordBytecodeCallTrace("call_member", memberName, "resolved_method", "inline", traceNode)
			}
			if vm.interp != nil && vm.interp.bytecodeStatsEnabled {
				vm.interp.recordBytecodeInlineCallHit()
			}
			if statsEnabled {
				vm.interp.recordBytecodeCallMemberDispatch(bytecodeCallMemberStatsResolvedInline)
			}
			return newProg, nil
		} else if vm.interp != nil && vm.interp.bytecodeStatsEnabled {
			vm.interp.recordBytecodeInlineCallMiss()
		}
		if !overloadDistinctReceiver {
			if result, handled, err := vm.tryCallResolvedCallableFromMemberStack(overloadFn, overloadReceiver, receiverIndex, argBase, argCount, callNode); handled || err != nil {
				if statsEnabled {
					vm.interp.recordBytecodeCallMemberDispatch(bytecodeCallMemberStatsResolvedGeneric)
				}
				return vm.finishCompletedCall(result, err, callNode, nil)
			}
		}
		vm.truncateStack(receiverIndex)
		if statsEnabled {
			vm.interp.recordBytecodeCallMemberDispatch(bytecodeCallMemberStatsResolvedFallback)
		}
		result, err := vm.callResolvedCallableWithInjectedReceiver(overloadFn, overloadReceiver, explicitArgs, callNode)
		return vm.finishCompletedCall(result, err, callNode, nil)
	}
	if newProg, err := vm.tryInlineMemberCallableFromStack(callee, methodReceiver, argBase, argCount, receiverIndex, callNode, currentProgram); err != nil {
		return nil, err
	} else if newProg != nil {
		if vm.interp != nil {
			vm.interp.recordBytecodeCallTrace("call_member", memberName, "resolved_method", "inline", traceNode)
		}
		if vm.interp != nil && vm.interp.bytecodeStatsEnabled {
			vm.interp.recordBytecodeInlineCallHit()
		}
		if statsEnabled {
			vm.interp.recordBytecodeCallMemberDispatch(bytecodeCallMemberStatsResolvedInline)
		}
		return newProg, nil
	} else if vm.interp != nil && vm.interp.bytecodeStatsEnabled {
		vm.interp.recordBytecodeInlineCallMiss()
	}
	if vm.interp != nil {
		vm.interp.recordBytecodeCallTrace("call_member", memberName, "resolved_method", "generic", traceNode)
	}
	if !distinctInjectedReceiver {
		if result, handled, err := vm.tryCallResolvedCallableFromMemberStack(callee, methodReceiver, receiverIndex, argBase, argCount, callNode); handled || err != nil {
			if statsEnabled {
				vm.interp.recordBytecodeCallMemberDispatch(bytecodeCallMemberStatsResolvedGeneric)
			}
			return vm.finishCompletedCall(result, err, callNode, nil)
		}
	}
	vm.truncateStack(receiverIndex)
	if statsEnabled {
		vm.interp.recordBytecodeCallMemberDispatch(bytecodeCallMemberStatsResolvedFallback)
	}
	result, err := vm.callResolvedCallableWithInjectedReceiver(callee, methodReceiver, explicitArgs, callNode)
	return vm.finishCompletedCall(result, err, callNode, nil)
}

func (vm *bytecodeVM) execCallMemberArrayGet(instr bytecodeInstruction, currentProgram *bytecodeProgram) (*bytecodeProgram, error) {
	if instr.name != "get" || instr.argCount != 1 || instr.safe {
		return vm.execCallMember(instr, currentProgram)
	}
	if vm.stackDepth() >= 2 {
		receiverIndex := vm.stackDepth() - 2
		argBase := receiverIndex + 1
		receiver := vm.stackValue(receiverIndex)
		arr, arrOK := receiver.(*runtime.ArrayValue)
		idx, idxOK := bytecodeArrayGetIndexI32(vm.stackValue(argBase))
		if arrOK && idxOK &&
			vm.canUseCanonicalArrayGetCallCacheForArray(arr) &&
			vm.lookupCachedCanonicalArrayGetCallForArray(currentProgram, vm.ip) {
			var callNode *ast.FunctionCall
			if instr.node != nil {
				if call, ok := instr.node.(*ast.FunctionCall); ok {
					callNode = call
				}
			}
			if newProg, handled, err := vm.finishArrayGetMemberFast(instr, arr, idx, receiverIndex, callNode); handled {
				return newProg, err
			}
		}
	}
	return vm.execCallMember(instr, currentProgram)
}

func (vm *bytecodeVM) execCallMemberNext(instr bytecodeInstruction, currentProgram *bytecodeProgram) (*bytecodeProgram, error) {
	if instr.name != "next" || instr.argCount != 0 || instr.safe {
		return vm.execCallMember(instr, currentProgram)
	}
	if vm.stackDepth() >= 1 {
		receiverIndex := vm.stackDepth() - 1
		var callNode *ast.FunctionCall
		if instr.node != nil {
			if call, ok := instr.node.(*ast.FunctionCall); ok {
				callNode = call
			}
		}
		if newProg, handled, err := vm.execIteratorNextCallMemberFast(instr, receiverIndex, callNode); handled {
			return newProg, err
		}
		if newProg, handled, err := vm.execCanonicalStringByteIteratorNextCallMemberFast(instr, receiverIndex, callNode); handled {
			return newProg, err
		}
		if newProg, handled, err := vm.execCanonicalStringCharIteratorNextCallMemberFast(instr, receiverIndex, callNode); handled {
			return newProg, err
		}
		if newProg, handled, err := vm.execCachedNextCallMemberFast(instr, receiverIndex, callNode, currentProgram); handled {
			return newProg, err
		}
	}
	return vm.execCallMember(instr, currentProgram)
}

func (vm *bytecodeVM) execCallMemberArrayNew(instr bytecodeInstruction, currentProgram *bytecodeProgram) (*bytecodeProgram, error) {
	if instr.name != "new" || instr.argCount != 0 || instr.safe {
		return vm.execCallMember(instr, currentProgram)
	}
	if vm.stackDepth() < 1 {
		return nil, fmt.Errorf("bytecode stack underflow")
	}
	receiverIndex := vm.stackDepth() - 1
	receiver := vm.stackValue(receiverIndex)
	if !bytecodeCanonicalArrayDefinitionReceiver(vm.interp, receiver) {
		return vm.execCallMember(instr, currentProgram)
	}
	var callNode *ast.FunctionCall
	if instr.node != nil {
		if call, ok := instr.node.(*ast.FunctionCall); ok {
			callNode = call
		}
	}
	if vm.lookupCachedCanonicalArrayNewCall(currentProgram, vm.ip, instr, receiver) {
		newProg, _, err := vm.finishStaticArrayNewMemberFast(instr, receiverIndex, callNode)
		return newProg, err
	}
	memberExpr := ast.Expression(ast.ID(instr.name))
	callee, err := vm.interp.memberAccessOnValueWithOptions(receiver, memberExpr, vm.env, true)
	if err != nil {
		return nil, vm.attachBytecodeRuntimeContext(err, callNode, nil)
	}
	callIP := vm.ip
	if newProg, handled, err := vm.execStaticArrayNewMemberFast(instr, receiver, callee, receiverIndex, callNode); handled {
		vm.storeCachedCanonicalArrayNewCall(currentProgram, callIP, instr, receiver)
		return newProg, err
	}
	return vm.execCallMember(instr, currentProgram)
}

func (vm *bytecodeVM) execCallMemberArraySlot(instr *bytecodeInstruction, currentProgram *bytecodeProgram) (*bytecodeProgram, error) {
	kind, ok := bytecodeArraySlotCallFastPathForInstruction(instr)
	if !ok {
		return vm.execCallMember(*instr, currentProgram)
	}
	statsEnabled := vm.interp != nil && vm.interp.bytecodeStatsEnabled
	if statsEnabled {
		vm.interp.recordBytecodeArrayMemberSlotLookup(kind)
	}
	if vm.stackDepth() < instr.argCount+1 {
		return nil, fmt.Errorf("bytecode stack underflow")
	}
	receiverIndex := vm.stackDepth() - instr.argCount - 1
	receiver := vm.stackValue(receiverIndex)
	arr, arrOK := receiver.(*runtime.ArrayValue)
	if !arrOK || arr == nil {
		if statsEnabled {
			vm.interp.recordBytecodeArrayMemberSlotFallback(bytecodeArrayMemberSlotFallbackReceiverMiss)
		}
		return vm.execCallMember(*instr, currentProgram)
	}

	argBase := receiverIndex + 1
	var callNode *ast.FunctionCall
	if instr.node != nil {
		if call, ok := instr.node.(*ast.FunctionCall); ok {
			callNode = call
		}
	}
	globalRevision, methodVersion, noRuntimeData := vm.noRuntimeDataGlobalAndMethodVersions()
	if noRuntimeData &&
		vm.lookupCachedCanonicalArraySlotCallForArrayValidatedWithVersions(currentProgram, vm.ip, kind, globalRevision, methodVersion) {
		if statsEnabled {
			vm.interp.recordBytecodeArrayMemberSlotCacheHit()
		}
		switch kind {
		case bytecodeMemberMethodFastPathArrayLen:
			if newProg, handled, err := vm.execArrayLenMemberFast(*instr, receiverIndex, callNode); handled {
				if statsEnabled {
					vm.interp.recordBytecodeArrayMemberSlotFastHit()
				}
				return newProg, err
			}
		case bytecodeMemberMethodFastPathArrayReadSlot:
			if newProg, handled, err := vm.finishArrayReadSlotMemberFast(instr.name, instr.argCount, instr.node, arr, receiverIndex, argBase, callNode); handled {
				if statsEnabled {
					vm.interp.recordBytecodeArrayMemberSlotFastHit()
				}
				return newProg, err
			}
		case bytecodeMemberMethodFastPathArrayWriteSlot:
			if newProg, handled, err := vm.finishArrayWriteSlotMemberFast(instr.name, instr.argCount, instr.node, arr, receiverIndex, argBase, callNode); handled {
				if statsEnabled {
					vm.interp.recordBytecodeArrayMemberSlotFastHit()
				}
				return newProg, err
			}
		case bytecodeMemberMethodFastPathArrayPush:
			if newProg, handled, err := vm.execArrayPushMemberFast(instr.name, instr.argCount, instr.node, receiverIndex, argBase, callNode); handled {
				if statsEnabled {
					vm.interp.recordBytecodeArrayMemberSlotFastHit()
				}
				return newProg, err
			}
		}
		if statsEnabled {
			vm.interp.recordBytecodeArrayMemberSlotFallback(bytecodeArrayMemberSlotFallbackFastPathMiss)
		}
		return vm.execCallMember(*instr, currentProgram)
	}
	if vm.canUseMemberMethodCacheForReceiver(instr.name, true, arr) {
		if cached, ok := vm.lookupCachedMemberMethodEntry(currentProgram, vm.ip, instr.name, true, arr); ok && cached.fastPath == kind {
			if newProg, handled, err := vm.execCallMemberFastPath(cached.fastPath, *instr, receiverIndex, argBase, callNode, currentProgram, arr); handled {
				if statsEnabled {
					vm.interp.recordBytecodeArrayMemberSlotCacheHit()
					vm.interp.recordBytecodeArrayMemberSlotFastHit()
				}
				return newProg, err
			}
			if statsEnabled {
				vm.interp.recordBytecodeArrayMemberSlotFallback(bytecodeArrayMemberSlotFallbackFastPathMiss)
			}
			return vm.execCallMember(*instr, currentProgram)
		}
	}
	if statsEnabled {
		vm.interp.recordBytecodeArrayMemberSlotFallback(bytecodeArrayMemberSlotFallbackCacheMiss)
	}
	return vm.execCallMember(*instr, currentProgram)
}

func (vm *bytecodeVM) execCallMemberFastPath(kind bytecodeMemberMethodFastPathKind, instr bytecodeInstruction, receiverIndex int, argBase int, callNode *ast.FunctionCall, currentProgram *bytecodeProgram, receiver runtime.Value) (*bytecodeProgram, bool, error) {
	if bytecodeMemberMethodFastPathIsArraySlot(kind) {
		vm.storeCachedCanonicalArraySlotCall(currentProgram, vm.ip, instr, receiver, kind)
	}
	return vm.execCachedMemberMethodFastPath(kind, instr, receiverIndex, argBase, callNode)
}

func (vm *bytecodeVM) execCallMemberStructCallableField(instr bytecodeInstruction, receiver *runtime.StructInstanceValue, receiverIndex int, argBase int, callNode *ast.FunctionCall, currentProgram *bytecodeProgram, traceNode ast.Node, statsEnabled bool) (*bytecodeProgram, bool, error) {
	if vm == nil || receiver == nil || instr.name == "" {
		return nil, false, nil
	}
	callee, ok := structNamedFieldValue(receiver, instr.name)
	if !ok || !isCallableRuntimeValue(callee) {
		return nil, false, nil
	}
	if target, ok := bytecodeResolveExactNativeCallTarget(callee, instr.argCount); ok {
		if vm.interp != nil {
			vm.interp.recordBytecodeCallTrace("call_member", instr.name, "member_access", "exact_native", traceNode)
		}
		args := vm.stackValuesFrom(argBase)
		vm.truncateStack(receiverIndex)
		newProg, finishErr := vm.execAndFinishExactNativeCall(target, args, callNode)
		return newProg, true, finishErr
	}
	if newProg, err := vm.tryInlineMemberCallableFromStack(callee, receiver, argBase, instr.argCount, receiverIndex, callNode, currentProgram); err != nil {
		return nil, true, err
	} else if newProg != nil {
		if vm.interp != nil {
			vm.interp.recordBytecodeCallTrace("call_member", instr.name, "member_access", "inline", traceNode)
		}
		if statsEnabled {
			vm.interp.recordBytecodeInlineCallHit()
		}
		return newProg, true, nil
	} else if statsEnabled {
		vm.interp.recordBytecodeInlineCallMiss()
	}
	args := vm.stackValuesFrom(argBase)
	vm.truncateStack(receiverIndex)
	needsStableArgsCopy := bytecodeCallTargetNeedsStableArgs(callee)
	if needsStableArgsCopy {
		var inline [bytecodeInlinePreparedCallArgStorage]runtime.Value
		if prepared, ok := bytecodePrepareCallArgsIntoBuffer(inline[:], args, true); ok {
			args = prepared
		} else {
			args = vm.prepareMaterializedCallArgs(args, true, bytecodeMaterializationRequiredDynamic, bytecodeMaterializationReasonInterfaceUnion)
		}
	} else {
		args = vm.prepareMaterializedCallArgs(args, false, bytecodeMaterializationRequiredDynamic, bytecodeMaterializationReasonInterfaceUnion)
	}
	if vm.interp != nil {
		vm.interp.recordBytecodeCallTrace("call_member", instr.name, "member_access", "generic", traceNode)
	}
	result, err := vm.callCallableValueMutable(callee, args, callNode)
	newProg, finishErr := vm.finishCompletedCall(result, err, callNode, nil)
	return newProg, true, finishErr
}

// execCallMember handles bytecodeOpCallMember for the common `obj.method(...)`
// syntax path without materializing an intermediate bound-method value.
func (vm *bytecodeVM) execCallMember(instr bytecodeInstruction, currentProgram *bytecodeProgram) (*bytecodeProgram, error) {
	if instr.argCount < 0 {
		return nil, fmt.Errorf("bytecode call-member arg count invalid")
	}
	if vm.stackDepth() < instr.argCount+1 {
		return nil, fmt.Errorf("bytecode stack underflow")
	}
	if instr.name == "" {
		return nil, fmt.Errorf("bytecode call-member missing member name")
	}
	receiverIndex := vm.stackDepth() - instr.argCount - 1
	argBase := receiverIndex + 1
	receiver := vm.stackValue(receiverIndex)
	if instr.safe && isNilRuntimeValue(receiver) {
		vm.truncateStack(receiverIndex)
		vm.appendStackValue(runtime.NilValue{})
		vm.ip++
		return nil, nil
	}
	receiver = vm.materializePrimitiveValue(bytecodeMaterializationRequiredDynamic, bytecodeMaterializationReasonInterfaceUnion, receiver)
	vm.setStackValue(receiverIndex, receiver)
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
	if interfaceValue, ok := receiver.(*runtime.InterfaceValue); ok && vm.interp.interfaceMethodReturnsSelf(interfaceValue, instr.name) {
		return vm.execInterfaceSelfReturnMember(instr, interfaceValue, receiverIndex, argBase, callNode)
	}
	statsEnabled := vm.interp != nil && vm.interp.bytecodeStatsEnabled
	useMethodCache := vm.canUseMemberMethodCacheForReceiver(instr.name, true, receiver)

	if newProg, handled, err := vm.execStaticCanonicalStructMemberFast(instr, receiver, receiverIndex, argBase, callNode); handled {
		return newProg, err
	}
	if structReceiver, ok := receiver.(*runtime.StructInstanceValue); ok {
		if newProg, handled, err := vm.execCallMemberStructCallableField(instr, structReceiver, receiverIndex, argBase, callNode, currentProgram, traceNode, statsEnabled); handled {
			return newProg, err
		}
	}

	if bytecodeCanDirectMemberCall(receiver, instr.name) {
		if vm.lookupCachedCanonicalArrayGetCall(currentProgram, vm.ip, instr, receiver) {
			if newProg, handled, err := vm.execArrayGetMemberFast(instr, receiverIndex, argBase, callNode); handled {
				return newProg, err
			}
		}
		if useMethodCache {
			if cached, ok := vm.lookupCachedMemberMethodEntry(currentProgram, vm.ip, instr.name, true, receiver); ok {
				if newProg, handled, err := vm.execCallMemberFastPath(cached.fastPath, instr, receiverIndex, argBase, callNode, currentProgram, receiver); handled {
					return newProg, err
				}
				if cached.template != nil {
					return vm.execCachedResolvedMemberCall(cached, instr.name, receiverIndex, argBase, instr.argCount, callNode, currentProgram)
				}
			}
		}
		receiverTypeHint := vm.interp.staticReceiverTypeForCall(callNode, vm.env)
		var callable runtime.Value
		var injectedReceiver runtime.Value
		var hasInjectedReceiver bool
		var found bool
		var err error
		callable, injectedReceiver, hasInjectedReceiver, found, err = vm.interp.resolveDirectCallMemberCallable(vm.env, receiver, instr.name, receiverTypeHint)
		if err != nil {
			return nil, vm.attachBytecodeRuntimeContext(err, callNode, nil)
		}
		if found {
			methodReceiver := receiver
			if hasInjectedReceiver {
				methodReceiver = injectedReceiver
			}
			if useMethodCache {
				cacheTemplate := callable
				cacheOK := false
				if _, ok := receiver.(*runtime.InterfaceValue); ok {
					cacheOK = true
				} else {
					if bound, boundOK := bindMemberMethodTemplate(methodReceiver, callable); boundOK {
						cacheTemplate = bound
						cacheOK = true
					}
				}
				if cacheOK {
					vm.storeCachedMemberMethod(currentProgram, vm.ip, instr.name, true, receiver, cacheTemplate)
				}
			}
			distinctInjectedReceiver := bytecodeDirectMemberCallHasDistinctInjectedReceiver(receiver, injectedReceiver, hasInjectedReceiver)
			if !distinctInjectedReceiver {
				if fn, ok := bytecodeResolvedMemberFastPathFunction(callable); ok {
					kind := vm.resolvedMemberMethodFastPath(instr.name, methodReceiver, fn)
					if newProg, handled, err := vm.execCallMemberFastPath(kind, instr, receiverIndex, argBase, callNode, currentProgram, receiver); handled {
						return newProg, err
					}
				}
			}
			callIP := vm.ip
			if !distinctInjectedReceiver {
				if newProg, handled, err := vm.execCanonicalArrayGetOverloadMemberFast(callable, instr, receiverIndex, argBase, callNode); handled {
					vm.storeCachedCanonicalArrayGetCall(currentProgram, callIP, instr, receiver)
					return newProg, err
				}
			}
			if overloadFn, overloadReceiver, ok, err := vm.resolveConcreteMemberOverload(callable, methodReceiver, vm.stackValuesFrom(argBase), callNode); err != nil {
				return nil, err
			} else if ok {
				if !distinctInjectedReceiver {
					kind := vm.resolvedMemberMethodFastPath(instr.name, overloadReceiver, overloadFn)
					if newProg, handled, err := vm.execCallMemberFastPath(kind, instr, receiverIndex, argBase, callNode, currentProgram, receiver); handled {
						return newProg, err
					}
				}
				if !distinctInjectedReceiver {
					if newProg, err := vm.tryInlineResolvedCallFromStack(overloadFn, overloadReceiver, true, argBase, instr.argCount, receiverIndex, callNode, currentProgram); err != nil {
						return nil, err
					} else if newProg != nil {
						if vm.interp != nil {
							vm.interp.recordBytecodeCallTrace("call_member", instr.name, "resolved_method", "inline", traceNode)
						}
						if statsEnabled {
							vm.interp.recordBytecodeInlineCallHit()
							vm.interp.recordBytecodeCallMemberDispatch(bytecodeCallMemberStatsResolvedInline)
						}
						return newProg, nil
					} else if statsEnabled {
						vm.interp.recordBytecodeInlineCallMiss()
					}
				}
				args := vm.stackValuesFrom(argBase)
				if vm.interp != nil {
					vm.interp.recordBytecodeCallTrace("call_member", instr.name, "resolved_method", "generic", traceNode)
				}
				if !distinctInjectedReceiver {
					if result, handled, err := vm.tryCallResolvedCallableFromMemberStack(overloadFn, overloadReceiver, receiverIndex, argBase, instr.argCount, callNode); handled || err != nil {
						if statsEnabled {
							vm.interp.recordBytecodeCallMemberDispatch(bytecodeCallMemberStatsResolvedGeneric)
						}
						return vm.finishCompletedCall(result, err, callNode, nil)
					}
				}
				vm.truncateStack(receiverIndex)
				if statsEnabled {
					vm.interp.recordBytecodeCallMemberDispatch(bytecodeCallMemberStatsResolvedFallback)
				}
				result, err := vm.callResolvedCallableWithInjectedReceiver(overloadFn, overloadReceiver, args, callNode)
				return vm.finishCompletedCall(result, err, callNode, nil)
			}
			if target, ok := bytecodeResolveExactInjectedNativeCallTarget(callable, methodReceiver, instr.argCount); ok {
				if vm.interp != nil {
					vm.interp.recordBytecodeCallTrace("call_member", instr.name, "resolved_method", "exact_native", traceNode)
				}
				if statsEnabled {
					vm.interp.recordBytecodeCallMemberDispatch(bytecodeCallMemberStatsResolvedExactNative)
				}
				args := vm.stackValuesFrom(argBase)
				vm.truncateStack(receiverIndex)
				return vm.execAndFinishExactNativeCall(target, args, callNode)
			}
			if !distinctInjectedReceiver {
				if newProg, err := vm.tryInlineMemberCallableFromStack(callable, methodReceiver, argBase, instr.argCount, receiverIndex, callNode, currentProgram); err != nil {
					return nil, err
				} else if newProg != nil {
					if vm.interp != nil {
						vm.interp.recordBytecodeCallTrace("call_member", instr.name, "resolved_method", "inline", traceNode)
					}
					if statsEnabled {
						vm.interp.recordBytecodeInlineCallHit()
						vm.interp.recordBytecodeCallMemberDispatch(bytecodeCallMemberStatsResolvedInline)
					}
					return newProg, nil
				} else if statsEnabled {
					vm.interp.recordBytecodeInlineCallMiss()
				}
			}
			args := vm.stackValuesFrom(argBase)
			if vm.interp != nil {
				vm.interp.recordBytecodeCallTrace("call_member", instr.name, "resolved_method", "generic", traceNode)
			}
			if !distinctInjectedReceiver {
				if result, handled, err := vm.tryCallResolvedCallableFromMemberStack(callable, methodReceiver, receiverIndex, argBase, instr.argCount, callNode); handled || err != nil {
					if statsEnabled {
						vm.interp.recordBytecodeCallMemberDispatch(bytecodeCallMemberStatsResolvedGeneric)
					}
					return vm.finishCompletedCall(result, err, callNode, nil)
				}
			}
			vm.truncateStack(receiverIndex)
			if statsEnabled {
				vm.interp.recordBytecodeCallMemberDispatch(bytecodeCallMemberStatsResolvedFallback)
			}
			result, err := vm.callResolvedCallableWithInjectedReceiver(callable, methodReceiver, args, callNode)
			return vm.finishCompletedCall(result, err, callNode, nil)
		}
	}

	if newProg, handled, err := vm.execCanonicalStringByteIteratorNextCallMemberFast(instr, receiverIndex, callNode); handled {
		return newProg, err
	}
	if newProg, handled, err := vm.execCanonicalStringCharIteratorNextCallMemberFast(instr, receiverIndex, callNode); handled {
		return newProg, err
	}
	if cached, ok := vm.lookupCachedStaticMemberCall(currentProgram, vm.ip, instr.name, instr.argCount, receiver); ok {
		return vm.execStaticMemberCallable(cached.callable, instr, receiverIndex, argBase, callNode, traceNode, currentProgram, statsEnabled)
	}
	if callee, found, err := vm.resolveDirectStaticMemberCallable(receiver, instr.name, instr.argCount); err != nil {
		return nil, vm.attachBytecodeRuntimeContext(err, callNode, nil)
	} else if found {
		vm.storeCachedStaticMemberCall(currentProgram, vm.ip, instr.name, instr.argCount, receiver, callee)
		return vm.execStaticMemberCallable(callee, instr, receiverIndex, argBase, callNode, traceNode, currentProgram, statsEnabled)
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
	if overloadFn, overloadReceiver, ok, err := vm.resolveConcreteMemberOverload(callee, receiver, vm.stackValuesFrom(argBase), callNode); err != nil {
		return nil, err
	} else if ok {
		kind := vm.resolvedMemberMethodFastPath(instr.name, overloadReceiver, overloadFn)
		if newProg, handled, err := vm.execCallMemberFastPath(kind, instr, receiverIndex, argBase, callNode, currentProgram, receiver); handled {
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
		args := vm.stackValuesFrom(argBase)
		if vm.interp != nil {
			vm.interp.recordBytecodeCallTrace("call_member", instr.name, "member_access", "generic", traceNode)
		}
		if result, handled, err := vm.tryCallResolvedCallableFromMemberStack(overloadFn, receiver, receiverIndex, argBase, instr.argCount, callNode); handled || err != nil {
			return vm.finishCompletedCall(result, err, callNode, nil)
		}
		vm.truncateStack(receiverIndex)
		result, err := vm.callResolvedCallableWithInjectedReceiver(overloadFn, overloadReceiver, args, callNode)
		return vm.finishCompletedCall(result, err, callNode, nil)
	}
	if target, ok := bytecodeResolveExactNativeCallTarget(callee, instr.argCount); ok {
		if vm.interp != nil {
			vm.interp.recordBytecodeCallTrace("call_member", instr.name, "member_access", "exact_native", traceNode)
		}
		args := vm.stackValuesFrom(argBase)
		vm.truncateStack(receiverIndex)
		return vm.execAndFinishExactNativeCall(target, args, callNode)
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
	args := vm.stackValuesFrom(argBase)
	vm.truncateStack(receiverIndex)
	needsStableArgsCopy := bytecodeCallTargetNeedsStableArgs(callee)
	if needsStableArgsCopy {
		var inline [bytecodeInlinePreparedCallArgStorage]runtime.Value
		if prepared, ok := bytecodePrepareCallArgsIntoBuffer(inline[:], args, true); ok {
			args = prepared
		} else {
			args = vm.prepareMaterializedCallArgs(args, true, bytecodeMaterializationRequiredDynamic, bytecodeMaterializationReasonInterfaceUnion)
		}
	} else {
		args = vm.prepareMaterializedCallArgs(args, false, bytecodeMaterializationRequiredDynamic, bytecodeMaterializationReasonInterfaceUnion)
	}
	if vm.interp != nil {
		vm.interp.recordBytecodeCallTrace("call_member", instr.name, "member_access", "generic", traceNode)
	}
	result, err := vm.callCallableValueMutable(callee, args, callNode)
	return vm.finishCompletedCall(result, err, callNode, nil)
}
