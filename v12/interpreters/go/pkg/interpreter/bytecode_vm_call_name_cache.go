package interpreter

import (
	"fmt"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

type bytecodeCallNameDispatchKind uint8

const (
	bytecodeCallNameDispatchGeneric bytecodeCallNameDispatchKind = iota
	bytecodeCallNameDispatchExactNative
	bytecodeCallNameDispatchInline
)

type bytecodeCallNameCacheEntry struct {
	name                 string
	env                  *runtime.Environment
	envVersion           uint64
	owner                *runtime.Environment
	ownerVersion         uint64
	callee               runtime.Value
	dispatch             bytecodeCallNameDispatchKind
	exactTarget          bytecodeExactNativeCallTarget
	inlineFn             *runtime.FunctionValue
	injectedReceiver     runtime.Value
	hasInjectedReceiver  bool
	inlineProgram        *bytecodeProgram
	inlineLayout         *bytecodeFrameLayout
	inlineReturnGenerics map[string]struct{}
	inlineDirect         bool
	needsStableArgsCopy  bool
}

type bytecodeInlineCallNameCacheEntry struct {
	valid   bool
	program *bytecodeProgram
	ip      int
	entry   *bytecodeCallNameCacheEntry
}

func bytecodeBuildCallNameCacheEntry(name string, lookup bytecodeResolvedIdentifierLookup, callee runtime.Value, argCount int, callNode *ast.FunctionCall) bytecodeCallNameCacheEntry {
	entry := bytecodeCallNameCacheEntry{
		name:                name,
		env:                 lookup.env,
		envVersion:          lookup.envVersion,
		owner:               lookup.owner,
		ownerVersion:        lookup.ownerVersion,
		callee:              callee,
		dispatch:            bytecodeCallNameDispatchGeneric,
		needsStableArgsCopy: bytecodeCallTargetNeedsStableArgs(callee),
	}
	if target, ok := bytecodeResolveExactNativeCallTarget(callee, argCount); ok {
		entry.dispatch = bytecodeCallNameDispatchExactNative
		entry.exactTarget = target
		return entry
	}
	if fn, injectedReceiver, hasInjectedReceiver, ok := inlineCallFunctionValue(callee); ok {
		entry.dispatch = bytecodeCallNameDispatchInline
		entry.inlineFn = fn
		entry.injectedReceiver = injectedReceiver
		entry.hasInjectedReceiver = hasInjectedReceiver
		if !hasInjectedReceiver {
			entry.inlineProgram, entry.inlineLayout, entry.inlineReturnGenerics, entry.inlineDirect = bytecodeDirectCallNameInlineShape(fn, argCount, callNode)
		}
	}
	return entry
}

func bytecodeDirectCallNameInlineShape(fn *runtime.FunctionValue, argCount int, callNode *ast.FunctionCall) (*bytecodeProgram, *bytecodeFrameLayout, map[string]struct{}, bool) {
	if fn == nil {
		return nil, nil, nil, false
	}
	if callNode != nil && len(callNode.TypeArguments) > 0 {
		return nil, nil, nil, false
	}
	if fn.MethodSet != nil && len(fn.MethodSet.GenericParams) > 0 {
		return nil, nil, nil, false
	}
	prog, ok := fn.Bytecode.(*bytecodeProgram)
	if !ok || prog == nil || prog.frameLayout == nil {
		return nil, nil, nil, false
	}
	layout := prog.frameLayout
	if layout.methodShorthand || layout.paramSlots != argCount {
		return nil, nil, nil, false
	}
	return prog, layout, bytecodeInlineReturnGenericNames(fn, prog), true
}

func (vm *bytecodeVM) lookupCachedCallName(program *bytecodeProgram, ip int, name string) (*bytecodeCallNameCacheEntry, bool) {
	if vm == nil || vm.env == nil {
		return nil, false
	}
	currentEnv := vm.env
	if hot := vm.callNameHot; hot.valid &&
		hot.program == program &&
		hot.ip == ip &&
		hot.entry != nil &&
		hot.entry.name == name &&
		hot.entry.env == currentEnv {
		entry := hot.entry
		currentEnvVersion := vm.bytecodeEnvRevision(currentEnv)
		if entry.envVersion != currentEnvVersion {
			return nil, false
		}
		if entry.owner == nil || entry.ownerVersion != vm.bytecodeEnvRevision(entry.owner) {
			return nil, false
		}
		return entry, true
	}
	if vm.callNameCache == nil {
		return nil, false
	}
	key := bytecodeGlobalLookupCacheKey{program: program, ip: ip}
	entry, ok := vm.callNameCache[key]
	if !ok || entry == nil || entry.name != name || entry.env != currentEnv {
		return nil, false
	}
	currentEnvVersion := vm.bytecodeEnvRevision(currentEnv)
	if entry.envVersion != currentEnvVersion {
		return nil, false
	}
	if entry.owner == nil || entry.ownerVersion != vm.bytecodeEnvRevision(entry.owner) {
		return nil, false
	}
	vm.callNameHot = bytecodeInlineCallNameCacheEntry{
		valid:   true,
		program: program,
		ip:      ip,
		entry:   entry,
	}
	return entry, true
}

func (vm *bytecodeVM) storeCachedCallName(program *bytecodeProgram, ip int, entry bytecodeCallNameCacheEntry) *bytecodeCallNameCacheEntry {
	if vm == nil || program == nil || entry.name == "" || entry.env == nil || entry.owner == nil {
		return nil
	}
	if vm.callNameCache == nil {
		vm.callNameCache = make(map[bytecodeGlobalLookupCacheKey]*bytecodeCallNameCacheEntry, 8)
	}
	key := bytecodeGlobalLookupCacheKey{program: program, ip: ip}
	cached := vm.callNameCache[key]
	if cached == nil {
		cached = new(bytecodeCallNameCacheEntry)
		vm.callNameCache[key] = cached
	}
	*cached = entry
	vm.callNameHot = bytecodeInlineCallNameCacheEntry{
		valid:   true,
		program: program,
		ip:      ip,
		entry:   cached,
	}
	return cached
}

func (vm *bytecodeVM) execCachedCallName(entry *bytecodeCallNameCacheEntry, argBase int, argCount int, callNode *ast.FunctionCall, currentProgram *bytecodeProgram) (*bytecodeProgram, error) {
	if entry == nil {
		return nil, fmt.Errorf("bytecode cached call entry missing")
	}
	traceEnabled := vm.interp != nil && vm.interp.bytecodeTraceEnabled
	var traceNode ast.Node
	if traceEnabled && callNode != nil {
		traceNode = callNode
	}
	switch entry.dispatch {
	case bytecodeCallNameDispatchExactNative:
		if traceEnabled {
			vm.interp.recordBytecodeCallTrace("call_name", entry.name, "name", "exact_native", traceNode)
		}
		args := vm.stack[argBase:]
		vm.stack = vm.stack[:argBase]
		result, _, err := vm.execExactNativeCall(entry.exactTarget, args, callNode)
		return vm.finishCompletedCall(result, err, callNode, nil)
	case bytecodeCallNameDispatchInline:
		if newProg, handled, err := vm.tryInlineCachedCallNameDirectFromStack(entry, argBase, argCount, callNode, currentProgram); handled || err != nil {
			if err != nil {
				return nil, err
			}
			if traceEnabled {
				vm.interp.recordBytecodeCallTrace("call_name", entry.name, "name", "inline", traceNode)
			}
			if vm.interp != nil && vm.interp.bytecodeStatsEnabled {
				vm.interp.recordBytecodeInlineCallHit()
			}
			return newProg, nil
		}
		if newProg, err := vm.tryInlineResolvedCallFromStack(entry.inlineFn, entry.injectedReceiver, entry.hasInjectedReceiver, argBase, argCount, argBase, callNode, currentProgram); err != nil {
			return nil, err
		} else if newProg != nil {
			if traceEnabled {
				vm.interp.recordBytecodeCallTrace("call_name", entry.name, "name", "inline", traceNode)
			}
			if vm.interp != nil && vm.interp.bytecodeStatsEnabled {
				vm.interp.recordBytecodeInlineCallHit()
			}
			return newProg, nil
		}
		if vm.interp != nil && vm.interp.bytecodeStatsEnabled {
			vm.interp.recordBytecodeInlineCallMiss()
		}
	}
	args := vm.stack[argBase:]
	vm.stack = vm.stack[:argBase]
	if entry.needsStableArgsCopy {
		args = copyCallArgs(args)
	}
	if traceEnabled {
		vm.interp.recordBytecodeCallTrace("call_name", entry.name, "name", "generic", traceNode)
	}
	result, err := vm.interp.callCallableValueMutable(entry.callee, args, vm.env, callNode)
	return vm.finishCompletedCall(result, err, callNode, nil)
}

func (vm *bytecodeVM) tryInlineCachedCallNameDirectFromStack(entry *bytecodeCallNameCacheEntry, argBase int, argCount int, callNode *ast.FunctionCall, currentProgram *bytecodeProgram) (*bytecodeProgram, bool, error) {
	if entry == nil || !entry.inlineDirect {
		return nil, false, nil
	}
	if argBase < 0 || argCount < 0 || argBase+argCount > len(vm.stack) {
		return nil, true, fmt.Errorf("bytecode stack underflow")
	}
	if callNode != nil && len(callNode.TypeArguments) > 0 {
		return nil, false, nil
	}
	fn := entry.inlineFn
	prog := entry.inlineProgram
	layout := entry.inlineLayout
	if fn == nil || prog == nil || layout == nil || argCount != layout.paramSlots {
		return nil, false, nil
	}
	slots := vm.acquireSlotFrame(layout.slotCount)
	if !layout.anyParamCoercion {
		inlineCopyArgsToSlots(slots, vm.stack[argBase:argBase+argCount], argCount)
	} else {
		for idx := 0; idx < layout.paramSlots; idx++ {
			arg := vm.stack[argBase+idx]
			paramType := inlineParamType(layout, idx)
			if inlineParamNeedsRuntimeCoercion(layout, idx, fn) && !inlineParamCoercionUnnecessary(layout, idx, paramType, arg) {
				if coerced, ok, err := inlineCoerceValueBySimpleType(inlineParamSimpleType(layout, idx), arg); err != nil {
					vm.releaseSlotFrame(slots)
					return nil, true, err
				} else if ok {
					arg = coerced
				} else {
					coerced, err := vm.interp.coerceValueToType(paramType, arg)
					if err != nil {
						vm.releaseSlotFrame(slots)
						return nil, true, err
					}
					arg = coerced
				}
			}
			slots[idx] = arg
		}
	}
	if layout.selfCallSlot >= 0 && layout.selfCallSlot < len(slots) {
		slots[layout.selfCallSlot] = fn
	}

	hasImplicit := layout.paramSlots > 0 && layout.usesImplicitMember
	if hasImplicit {
		state := vm.interp.stateFromEnv(fn.Closure)
		state.pushImplicitReceiver(vm.stack[argBase])
	}

	vm.stack = vm.stack[:argBase]
	selfFast := bytecodeCanUseSelfFastFrame(currentProgram, prog, vm.env, fn.Closure)
	vm.pushCallFrame(vm.ip+1, currentProgram, vm.slots, vm.env, entry.inlineReturnGenerics, len(vm.iterStack), len(vm.loopStack), hasImplicit, selfFast)
	vm.slots = slots
	vm.env = fn.Closure
	vm.ip = 0
	return prog, true, nil
}
