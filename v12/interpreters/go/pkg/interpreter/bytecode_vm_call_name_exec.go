package interpreter

import (
	"fmt"
	"strings"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

// execCallName handles bytecodeOpCallName. It returns a non-nil program when
// an inline call frame was set up (the caller must switch to the new program).
// A nil program with nil error means the call completed normally.
func (vm *bytecodeVM) execCallName(instr bytecodeInstruction, currentProgram *bytecodeProgram) (*bytecodeProgram, error) {
	if instr.argCount < 0 {
		return nil, fmt.Errorf("bytecode call arg count invalid")
	}
	if instr.name == "" {
		return nil, fmt.Errorf("bytecode call missing target name")
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
	traceLookup := "name"
	statsEnabled := vm.interp != nil && vm.interp.bytecodeStatsEnabled
	traceEnabled := vm.interp != nil && vm.interp.bytecodeTraceEnabled
	recordInlineCallNameHit := func(kind bytecodeCallNameDispatchStats) {
		if traceEnabled {
			vm.interp.recordBytecodeCallTrace("call_name", instr.name, traceLookup, "inline", traceNode)
		}
		if statsEnabled {
			vm.interp.recordBytecodeCallNameDispatch(kind)
			vm.interp.recordBytecodeInlineCallHit()
		}
	}
	if !instr.slotArgs && vm.stackDepth() < instr.argCount {
		return nil, fmt.Errorf("bytecode stack underflow")
	}
	if statsEnabled {
		vm.interp.recordBytecodeCallNameLookup()
	}
	if instr.nameSimple {
		if cached, ok := vm.lookupCachedCallName(currentProgram, vm.ip, instr.name); ok {
			if instr.slotArgs {
				if newProg, handled, err := vm.tryInlineCachedCallNameDirectFromSlots(cached, instr, callNode, currentProgram); handled || err != nil {
					if err != nil {
						return nil, err
					}
					recordInlineCallNameHit(bytecodeCallNameStatsInlineDirectSlot)
					return newProg, nil
				}
				if err := vm.pushCallNameSlotArgs(instr); err != nil {
					return nil, err
				}
			}
			argBase := vm.stackDepth() - instr.argCount
			return vm.execCachedCallName(cached, argBase, instr.argCount, callNode, currentProgram)
		}
	}
	var (
		calleeVal runtime.Value
		found     bool
		lookup    bytecodeResolvedIdentifierLookup
	)
	if instr.nameSimple {
		lookup, found = vm.lookupIdentifierNameForCallCache(currentProgram, vm.ip, instr.name)
		calleeVal = lookup.value
	} else {
		calleeVal, found = vm.lookupCachedName(currentProgram, vm.ip, instr.name)
	}
	if !found {
		if !instr.nameSimple {
			dotIdx := strings.Index(instr.name, ".")
			if dotIdx <= 0 || dotIdx >= len(instr.name)-1 {
				err := fmt.Errorf("Undefined variable '%s'", instr.name)
				return nil, vm.attachBytecodeRuntimeContext(err, callNode, nil)
			}
			traceLookup = "dot_fallback"
			if statsEnabled {
				vm.interp.recordBytecodeCallNameDotFallback()
			}
			head := instr.name[:dotIdx]
			tail := instr.name[dotIdx+1:]
			receiver, recvFound := vm.lookupCachedName(currentProgram, vm.ip, head)
			if !recvFound {
				if def, ok := vm.env.StructDefinition(head); ok {
					receiver = def
				} else {
					receiver = runtime.TypeRefValue{TypeName: head}
				}
			}
			if cached, ok := vm.lookupCachedMemberMethod(currentProgram, vm.ip, tail, true, receiver); ok {
				calleeVal = cached
			} else {
				member := ast.ID(tail)
				candidate, err := vm.interp.memberAccessOnValueWithOptions(receiver, member, vm.env, true)
				if err != nil {
					return nil, vm.attachBytecodeRuntimeContext(err, callNode, nil)
				}
				calleeVal = candidate
				vm.storeCachedMemberMethod(currentProgram, vm.ip, tail, true, receiver, candidate)
			}
		} else {
			err := fmt.Errorf("Undefined variable '%s'", instr.name)
			return nil, vm.attachBytecodeRuntimeContext(err, callNode, nil)
		}
	}
	if instr.nameSimple && found {
		entry := bytecodeBuildCallNameCacheEntry(instr.name, lookup, calleeVal, instr.argCount, callNode)
		if cached := vm.storeCachedCallName(currentProgram, vm.ip, entry); cached != nil {
			if instr.slotArgs {
				if newProg, handled, err := vm.tryInlineCachedCallNameDirectFromSlots(cached, instr, callNode, currentProgram); handled || err != nil {
					if err != nil {
						return nil, err
					}
					recordInlineCallNameHit(bytecodeCallNameStatsInlineDirectSlot)
					return newProg, nil
				}
				if err := vm.pushCallNameSlotArgs(instr); err != nil {
					return nil, err
				}
			}
			argBase := vm.stackDepth() - instr.argCount
			return vm.execCachedCallName(cached, argBase, instr.argCount, callNode, currentProgram)
		}
		return nil, fmt.Errorf("bytecode call-name cache store failed")
	}
	if instr.slotArgs {
		if err := vm.pushCallNameSlotArgs(instr); err != nil {
			return nil, err
		}
	}
	argBase := vm.stackDepth() - instr.argCount
	if target, ok := bytecodeResolveExactNativeCallTarget(calleeVal, instr.argCount); ok {
		if traceEnabled {
			vm.interp.recordBytecodeCallTrace("call_name", instr.name, traceLookup, "exact_native", traceNode)
		}
		if statsEnabled {
			vm.interp.recordBytecodeCallNameDispatch(bytecodeCallNameStatsExactNative)
		}
		args := vm.stackValuesFrom(argBase)
		vm.truncateStack(argBase)
		return vm.execAndFinishExactNativeCall(target, args, callNode)
	}
	if newProg, err := vm.tryInlineCallFromStack(calleeVal, argBase, instr.argCount, argBase, callNode, currentProgram); err != nil {
		return nil, err
	} else if newProg != nil {
		recordInlineCallNameHit(bytecodeCallNameStatsInlineGeneric)
		return newProg, nil
	} else if statsEnabled {
		vm.interp.recordBytecodeInlineCallMiss()
	}
	args := vm.stackValuesFrom(argBase)
	vm.truncateStack(argBase)
	needsStableArgsCopy := bytecodeCallTargetNeedsStableArgs(calleeVal)
	if needsStableArgsCopy {
		var inline [bytecodeInlinePreparedCallArgStorage]runtime.Value
		if prepared, ok := bytecodePrepareCallArgsIntoBuffer(inline[:], args, true); ok {
			args = prepared
		} else {
			args = vm.prepareMaterializedCallArgs(args, true, bytecodeMaterializationCandidateStatic, bytecodeMaterializationReasonStaticCall)
		}
	} else {
		args = vm.prepareMaterializedCallArgs(args, false, bytecodeMaterializationCandidateStatic, bytecodeMaterializationReasonStaticCall)
	}
	if traceEnabled {
		vm.interp.recordBytecodeCallTrace("call_name", instr.name, traceLookup, "generic", traceNode)
	}
	if statsEnabled {
		vm.interp.recordBytecodeCallNameDispatch(bytecodeCallNameStatsGenericFallback)
	}
	result, err := vm.callCallableValueMutable(calleeVal, args, callNode)
	return vm.finishCompletedCall(result, err, callNode, nil)
}
